package main

import (
	"bytes"
	"encoding/gob"
)

// Constantes de tipos de mensaje
const (
	ANNOUNCE         = iota // 0
	REQUEST_PIECE           // 1
	ASSIGN_PORT             // 2
	REGISTER_TORRENT        // 3
	GET_STATE               // 4
	STATE_RESPONSE          // 5
	PIECE_DATA              // 6
)

// Message define la estructura de los mensajes de la red
type Message struct {
	Type    int
	Payload []byte
}

// Encode convierte el mensaje a bytes para enviarlo por el socket
func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}
