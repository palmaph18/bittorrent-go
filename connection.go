package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// ConnectionState maneja la información y el socket TCP de un peer conectado
type ConnectionState struct {
	Conn       net.Conn
	RemoteAddr string
	Bitfield   []bool // True si el peer tiene la pieza en ese índice
	Active     bool
	mu         sync.Mutex
}

// TransferProgress maneja la tolerancia a fallos guardando el estado local
type TransferProgress struct {
	TorrentHash    string  `json:"torrent_hash"`
	Downloaded     []bool  `json:"downloaded_pieces"` // Piezas completadas
	ProgressStatus float64 `json:"progress_status"`   // Porcentaje exacto de progreso
}

// NewConnectionState envuelve una conexión neta con su estado P2P
func NewConnectionState(conn net.Conn, numPieces int) *ConnectionState {
	return &ConnectionState{
		Conn:       conn,
		RemoteAddr: conn.RemoteAddr().String(),
		Bitfield:   make([]bool, numPieces),
		Active:     true,
	}
}

// Send envía un mensaje serializado a través del socket TCP
func (cs *ConnectionState) Send(msg *Message) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	data, err := msg.Encode() // Serialización a bytes definida en message.go
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %v", err)
	}

	_, err = cs.Conn.Write(data)
	return err
}

// Close cierra la conexión limpiamente y actualiza el estado
func (cs *ConnectionState) Close() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.Active {
		cs.Active = false
		cs.Conn.Close()
		fmt.Printf("Conexión cerrada con %s\n", cs.RemoteAddr)
	}
}

// SaveProgress persiste el progreso en disco para tolerancia a fallos
func SaveProgress(hash string, downloaded []bool, totalPieces int) error {
	completed := 0
	for _, done := range downloaded {
		if done {
			completed++
		}
	}

	percentage := (float64(completed) / float64(totalPieces)) * 100.0

	progress := &TransferProgress{
		TorrentHash:    hash,
		Downloaded:     downloaded,
		ProgressStatus: percentage,
	}

	path := filepath.Join("torrents", fmt.Sprintf("%s.progress", hash))
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(progress)
}

// LoadProgress recupera el estado de descarga tras una desconexión
func LoadProgress(hash string, totalPieces int) (*TransferProgress, error) {
	path := filepath.Join("torrents", fmt.Sprintf("%s.progress", hash))

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		// Si no existe, es una descarga nueva; empezamos de cero
		return &TransferProgress{
			TorrentHash:    hash,
			Downloaded:     make([]bool, totalPieces),
			ProgressStatus: 0.0,
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var progress TransferProgress
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&progress); err != nil {
		return nil, err
	}

	return &progress, nil
}
