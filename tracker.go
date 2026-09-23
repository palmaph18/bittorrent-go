package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Tracker gestiona el estado global y las solicitudes de conexión
type Tracker struct {
	TrackMap     map[string]*TorrentTrack // Mapa con el estado global
	Port         int
	mu           sync.RWMutex
	activePorts  map[int]bool // Administra los puertos de la red
	lastAssigned int          // Para asignar puertos de 1000 en 1000
}

// NewTracker inicializa el servidor centralizado
func NewTracker() *Tracker {
	return &Tracker{
		TrackMap:     make(map[string]*TorrentTrack),
		Port:         5000, // Operará por defecto en el puerto 5000
		activePorts:  make(map[int]bool),
		lastAssigned: 3000, // Iniciaremos sumando 1000 para que el primero sea 4000
	}
}

// AssignPort otorga un puerto a un nuevo Peer (ej. 4000, 6000, 7000)
func (t *Tracker) AssignPort() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	for {
		t.lastAssigned += 1000
		if t.lastAssigned == 5000 { // Evitar colisiones con el puerto del Tracker
			continue
		}
		if !t.activePorts[t.lastAssigned] {
			t.activePorts[t.lastAssigned] = true
			return t.lastAssigned
		}
	}
}

// ReleasePeer libera el puerto y remueve al nodo cuando se desconecta
func (t *Tracker) ReleasePeer(port int, peerIP string, torrentHash string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Mecanismo para limpiar puertos
	delete(t.activePorts, port)

	// Limpiar el nodo del mapa global
	if track, exists := t.TrackMap[torrentHash]; exists {
		delete(track.Peers, peerIP)
	}
	log.Printf("Nodo en %s:%d desconectado. Puerto liberado.", peerIP, port)
}

// Start iniciará el servidor (Sockets TCP)
func (t *Tracker) Start() {
	address := fmt.Sprintf(":%d", t.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error al iniciar el Tracker: %v\n", err)
	}
	defer listener.Close()

	fmt.Printf("\n=== Tracker Iniciado en el puerto %d ===\n", t.Port)

	// Goroutine que preguntará periódicamente el estado de los nodos
	go t.monitorNodes()

	// Goroutine para la opción obligatoria de CLI: ver estado global de la red
	go t.cliDashboard()

	// Bucle para gestionar solicitudes de conexión
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error aceptando conexión de peer: %v", err)
			continue
		}
		go t.handleConnection(conn)
	}
}

// handleConnection maneja la conexión entrante de un nodo
func (t *Tracker) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Usamos gob para decodificar el mensaje entrante
	decoder := gob.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return // Ignorar si el peer se desconecta
	}

	// Evaluamos el tipo de mensaje usando un switch (buenas prácticas en Go)
	switch msg.Type {
	case ANNOUNCE:
		port := t.AssignPort() // Administra los puertos de la red (ej. 4000, 6000)

		// Responder enviando el puerto asignado
		portStr := fmt.Sprintf("%d", port)
		resp := Message{Type: ASSIGN_PORT, Payload: []byte(portStr)}

		encoder := gob.NewEncoder(conn)
		encoder.Encode(&resp)

		fmt.Printf("\n[*] Nuevo Peer registrado. Puerto asignado: %d\n", port)

	case REGISTER_TORRENT:
		// Extraer datos del payload separados por '|'
		parts := strings.Split(string(msg.Payload), "|")
		if len(parts) == 4 {
			torrentHash := parts[0]
			var port int
			var dlStatus float64

			fmt.Sscanf(parts[1], "%d", &port)
			role := parts[2]
			fmt.Sscanf(parts[3], "%f", &dlStatus)

			// Obtener IP real
			peerIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()

			// Escribir en el mapa global con sincronización
			t.mu.Lock()
			if _, exists := t.TrackMap[torrentHash]; !exists {
				t.TrackMap[torrentHash] = &TorrentTrack{Peers: make(map[string]*PeerInfo)}
			}
			t.TrackMap[torrentHash].Peers[peerIP] = &PeerInfo{
				IP:             peerIP,
				Port:           port,
				Role:           role,
				DownloadStatus: dlStatus,
			}
			t.mu.Unlock()

			fmt.Printf("\n[*] Nodo %s:%d registrado en '%s' como %s\n", peerIP, port, torrentHash, role)
		}

	case GET_STATE:
		t.mu.RLock()
		// Convertimos el TrackMap a JSON para enviarlo fácilmente por la red
		stateBytes, _ := json.Marshal(t.TrackMap)
		t.mu.RUnlock()

		resp := Message{Type: STATE_RESPONSE, Payload: stateBytes}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(&resp)
	}
}

// monitorNodes hace un chequeo de salud y estado a los nodos
func (t *Tracker) monitorNodes() {
	for {
		time.Sleep(15 * time.Second) // Preguntará periódicamente
		t.mu.RLock()
		// Aquí iteraremos sobre TrackMap para enviar pings a los Peers

		// Lectura temporal del mapa para quitar el warning de "sección vacía"
		_ = len(t.TrackMap)

		// y actualizar sus porcentajes o aplicar Tolerancia a Fallos
		t.mu.RUnlock()
	}
}

// cliDashboard muestra la lista de nodos, roles, archivos y progresos
func (t *Tracker) cliDashboard() {
	for {
		time.Sleep(10 * time.Second) // Refrescar el dashboard cada 10 segundos

		t.mu.RLock()
		fmt.Println("\n--- ESTADO GLOBAL DE LA RED ---")
		if len(t.TrackMap) == 0 {
			fmt.Println("No hay torrents activos en la red.")
		} else {
			for hash, track := range t.TrackMap {
				fmt.Printf("Torrent [Hash: %s]\n", hash)
				for ip, peer := range track.Peers {
					fmt.Printf(" -> Nodo [%s:%d] | Rol: %s | Progreso: %.2f%%\n",
						ip, peer.Port, peer.Role, peer.DownloadStatus)
				}
			}
		}
		fmt.Println("-------------------------------")
		t.mu.RUnlock()
	}
}
