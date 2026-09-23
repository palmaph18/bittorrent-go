package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Peer representa la lógica principal del nodo
type Peer struct {
	IP             string
	Port           int
	Role           string // Puede ser "Leecher" o "Seeder"
	DownloadStatus float64
	TrackerAddr    string
	mu             sync.RWMutex
}

// NewPeer inicializa un nuevo nodo
func NewPeer(trackerAddr string) *Peer {
	return &Peer{
		Role:           "Leecher", // Inician como Leechers por defecto si solo descargan
		DownloadStatus: 0.0,
		TrackerAddr:    trackerAddr,
	}
}

// Start arranca el nodo actuando simultáneamente como cliente y servidor
func (p *Peer) Start() {
	fmt.Printf("Conectando al Tracker en %s...\n", p.TrackerAddr)

	// Nos conectamos al Tracker
	conn, err := net.Dial("tcp", p.TrackerAddr)
	if err != nil {
		log.Fatalf("No se pudo conectar al Tracker: %v", err)
	}

	// Enviamos el mensaje ANNOUNCE
	msg := Message{Type: ANNOUNCE, Payload: []byte("HELLO")}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(&msg)

	// Esperamos la respuesta con el puerto asignado
	decoder := gob.NewDecoder(conn)
	var resp Message
	if err := decoder.Decode(&resp); err == nil && resp.Type == ASSIGN_PORT {
		// Convertimos el payload (bytes) a int para p.Port
		fmt.Sscanf(string(resp.Payload), "%d", &p.Port)
		fmt.Printf(" Tracker nos ha asignado el puerto TCP: %d\n", p.Port)
	}
	conn.Close() // Cerramos esta conexión inicial (handshake)

	// 1. Lanzar el servidor en una goroutine para gestionar cargas concurrentes
	go p.startServer()

	// 2. Iniciar el menú CLI interactivo en el hilo principal
	p.cliMenu()
}

// startServer escucha peticiones TCP entrantes de otros peers
func (p *Peer) startServer() {
	address := fmt.Sprintf(":%d", p.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error al iniciar servidor del Peer: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// Lanzar múltiples goroutines para gestionar cargas de pedazos concurrentes
		go p.handleConnection(conn)
	}
}

// handleConnection procesa peticiones P2P directas de otros nodos (como servidor)
func (p *Peer) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	// Bucle infinito para mantener la conexión TCP abierta mientras el Leecher pida pedazos
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return // Sale del ciclo si el Leecher cierra la conexión al terminar
		}

		if msg.Type == REQUEST_PIECE {
			// Extraer nombre del archivo y número de pedazo del payload
			parts := strings.Split(string(msg.Payload), "|")
			if len(parts) == 2 {
				fileName := parts[0]
				var pieceIndex int
				fmt.Sscanf(parts[1], "%d", &pieceIndex)

				// Abrir el archivo original
				sourcePath := filepath.Join("archivos", fileName)
				file, err := os.Open(sourcePath)
				if err != nil {
					fmt.Printf("[ERROR Seeder] No se pudo abrir %s\n", sourcePath)
					continue
				}

				// Calcular el offset exacto y posicionar el cursor de lectura
				offset := int64(pieceIndex * PieceSize) // PieceSize = 102400 (definido en torrent.go)
				file.Seek(offset, 0)

				// Leer únicamente el bloque correspondiente
				buffer := make([]byte, PieceSize)
				bytesRead, _ := file.Read(buffer)
				file.Close()

				// Enviar los bytes leídos (usando [:bytesRead] por si el último pedazo es menor a 102400)
				resp := Message{Type: PIECE_DATA, Payload: buffer[:bytesRead]}
				encoder.Encode(&resp)
			}
		}
	}
}

// UpdateProgress actualiza el progreso y aplica las reglas de estado dinámico
func (p *Peer) UpdateProgress(newProgress float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.DownloadStatus = newProgress

	// Regla de distribución: compartir automáticamente al superar el 20% de descarga
	if p.DownloadStatus > 20.0 && p.Role == "Leecher" {
		fmt.Println("\n[*] Progreso superó el 20%. Comenzando a compartir fragmentos hacia la red...")
		// Aquí se activaría la lógica para permitir cargas (uploads).
	}

	// Roles dinámicos: cambia a Seeder si posee el archivo completo
	if p.DownloadStatus == 100.0 && p.Role != "Seeder" {
		p.Role = "Seeder"
		fmt.Println("\n[*] Descarga completa al 100%. Rol cambiado dinámicamente a Seeder.")
	}
}

// cliMenu despliega el menú interactivo para el nodo
func (p *Peer) cliMenu() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- MENÚ CLI DEL NODO ---")
		fmt.Println("1. Agregar un torrent (iniciar descarga o compartir)")
		fmt.Println("2. Debuggeo (mostrar logs de conexión)")
		fmt.Println("3. Ver progreso local de las descargas (porcentaje o barra)")
		fmt.Println("4. Ver estado global de la red")
		fmt.Print("Selecciona una opción: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Print(">> Ingresa el nombre del archivo o torrent (ej. prueba.txt o prueba.txt.torrent): ")
			fileName, _ := reader.ReadString('\n')
			fileName = strings.TrimSpace(fileName)

			// --- LÓGICA DE LEECHER (DESCARGA) ---
			if strings.HasSuffix(fileName, ".torrent") {
				fmt.Println(">> Cargando metadatos para iniciar descarga P2P...")
				torrentPath := filepath.Join("torrents", fileName)
				metadata, err := LoadTorrentMetadata(torrentPath)
				if err != nil {
					fmt.Printf(">> [ERROR] No se pudo leer el .torrent: %v\n", err)
					continue
				}
				// Lanzar la descarga en una goroutine para no bloquear el menú CLI
				go p.startDownload(metadata)
				continue
			}

			// --- LÓGICA DE SEEDER (COMPARTIR) ---
			sourcePath := filepath.Join("archivos", fileName)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				fmt.Println(">> [ERROR] Archivo no encontrado. Asegúrate de colocarlo en /archivos.")
				continue
			}

			fmt.Println(">> Generando metadatos y binarizando (esto tomará un momento)...")
			metadata, err := GenerateTorrent(sourcePath, "torrents")
			if err != nil {
				fmt.Printf(">> [ERROR] No se pudo generar el torrent: %v\n", err)
				continue
			}

			p.UpdateProgress(100.0)
			fmt.Printf(">> ¡Torrent '%s.torrent' creado con éxito en /torrents!\n", metadata.FileName)

			// Registrarse en el Tracker global como Seeder
			connTracker, errTracker := net.Dial("tcp", p.TrackerAddr)
			if errTracker == nil {
				payload := fmt.Sprintf("%s|%d|%s|%.2f", metadata.FileName, p.Port, p.Role, p.DownloadStatus)
				msg := Message{Type: REGISTER_TORRENT, Payload: []byte(payload)}
				encoder := gob.NewEncoder(connTracker)
				encoder.Encode(&msg)
				connTracker.Close()
				fmt.Println(">> Registrado exitosamente en el Tracker global.")
			}
		case "2":
			fmt.Println(">> [DEBUG] Mostrando conexiones activas...")
			// Listar sockets de connection.go
		case "3":
			p.mu.RLock()
			fmt.Printf(">> Progreso local: %.2f%% | Rol actual: %s\n", p.DownloadStatus, p.Role)
			p.mu.RUnlock()
		case "4":
			fmt.Println(">> Solicitando estado global al Tracker...")
			connTracker, err := net.Dial("tcp", p.TrackerAddr)
			if err == nil {
				msg := Message{Type: GET_STATE, Payload: []byte{}}
				encoder := gob.NewEncoder(connTracker)
				encoder.Encode(&msg)

				decoder := gob.NewDecoder(connTracker)
				var resp Message
				if err := decoder.Decode(&resp); err == nil && resp.Type == STATE_RESPONSE {
					fmt.Println("\n--- ESTADO GLOBAL (VISTA DEL NODO) ---")
					var trackMap map[string]*TorrentTrack
					json.Unmarshal(resp.Payload, &trackMap)

					if len(trackMap) == 0 {
						fmt.Println("No hay torrents activos en la red.")
					} else {
						for hash, track := range trackMap {
							fmt.Printf("Torrent disponible: %s\n", hash)
							for ip, peerInfo := range track.Peers {
								fmt.Printf(" -> Nodo [%s:%d] | Rol: %s | Progreso: %.2f%%\n",
									ip, peerInfo.Port, peerInfo.Role, peerInfo.DownloadStatus)
							}
						}
					}
					fmt.Println("--------------------------------------")
				}
				connTracker.Close()
			} else {
				fmt.Println(">> [ERROR] No se pudo contactar al Tracker.")
			}
		default:
			fmt.Println("Opción inválida. Intenta de nuevo.")
		}
	}
}

// startDownload conecta directamente con un Seeder, solicita los pedazos y los ensambla en disco
func (p *Peer) startDownload(metadata *TorrentMetadata) {
	fmt.Printf("\n[*] Iniciando descarga binaria del archivo: %s\n", metadata.FileName)
	seederAddr := "127.0.0.1:4000" // Mantendremos el Seeder local para esta prueba

	// Crear el archivo de destino localmente agregando un prefijo para no sobrescribir el original
	destPath := filepath.Join("archivos", "descargado_"+metadata.FileName)
	file, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("[ERROR] No se pudo crear el archivo local: %v\n", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", seederAddr)
	if err != nil {
		fmt.Printf("[ERROR] No se pudo conectar al Seeder P2P: %v\n", err)
		return
	}
	defer conn.Close()

	totalPieces := len(metadata.PiecesHashes)
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for i := 0; i < totalPieces; i++ {
		// El payload incluye el nombre del archivo y el índice del pedazo (ej. "archivo.mp4|3")
		payload := fmt.Sprintf("%s|%d", metadata.FileName, i)
		req := Message{Type: REQUEST_PIECE, Payload: []byte(payload)}

		if err := encoder.Encode(&req); err != nil {
			fmt.Printf("[ERROR] Fallo al solicitar pedazo %d: %v\n", i, err)
			break
		}

		var resp Message
		if err := decoder.Decode(&resp); err == nil && resp.Type == PIECE_DATA {
			// Escribir los bytes recibidos en el offset correcto del archivo
			offset := int64(i * metadata.PieceLength)
			file.WriteAt(resp.Payload, offset)

			// Actualizar progreso dinámicamente y aplicar regla de distribución
			progress := (float64(i+1) / float64(totalPieces)) * 100
			p.UpdateProgress(progress)
		}
	}
	fmt.Printf("\n[*] ¡Descarga P2P completada! Archivo guardado en: %s\n", destPath)
}
