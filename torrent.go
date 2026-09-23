package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json" // Manejo de metadatos JSON
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PieceSize define el tamaño sugerido de pedazo: 102400 bytes
const PieceSize = 102400

// TorrentMetadata representa la estructura del archivo .torrent
type TorrentMetadata struct {
	FileName     string   `json:"file_name"`
	FileSize     int64    `json:"file_size"`
	PieceLength  int      `json:"piece_length"`
	PiecesHashes []string `json:"pieces_hashes"` // Arreglo de hashes MD5 para validación de pedazos
}

// GenerateTorrent procesa un archivo, aplica binarización y calcula los metadatos
func GenerateTorrent(sourceFilePath string, destDirectory string) (*TorrentMetadata, error) {
	file, err := os.Open(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo fuente: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error al leer información del archivo: %v", err)
	}

	metadata := &TorrentMetadata{
		FileName:    filepath.Base(sourceFilePath),
		FileSize:    fileInfo.Size(),
		PieceLength: PieceSize,
	}

	// Buffer para la binarización por pedazos de 102400 bytes
	buffer := make([]byte, PieceSize)

	for {
		bytesRead, err := file.Read(buffer)
		if bytesRead > 0 {
			// Cálculo de hashes MD5 por cada pedazo leído
			hash := md5.Sum(buffer[:bytesRead])
			hashString := hex.EncodeToString(hash[:])
			metadata.PiecesHashes = append(metadata.PiecesHashes, hashString)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error durante la binarización del archivo: %v", err)
		}
	}

	// Guardar el archivo .torrent generado en el directorio /torrents
	torrentFileName := filepath.Join(destDirectory, metadata.FileName+".torrent")
	err = SaveTorrentMetadata(metadata, torrentFileName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Torrent generado exitosamente en: %s\n", torrentFileName)
	fmt.Printf("Total de pedazos (102400 bytes c/u): %d\n", len(metadata.PiecesHashes))
	return metadata, nil
}

// SaveTorrentMetadata serializa la estructura a JSON y la guarda en disco usando encoding/json
func SaveTorrentMetadata(metadata *TorrentMetadata, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error al crear el archivo .torrent: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Formato legible para propósitos de depuración
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("error al codificar JSON: %v", err)
	}

	return nil
}

// LoadTorrentMetadata lee un archivo .torrent y lo deserializa desde JSON
func LoadTorrentMetadata(torrentPath string) (*TorrentMetadata, error) {
	file, err := os.Open(torrentPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo .torrent: %v", err)
	}
	defer file.Close()

	var metadata TorrentMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("error al decodificar JSON: %v", err)
	}

	return &metadata, nil
}
