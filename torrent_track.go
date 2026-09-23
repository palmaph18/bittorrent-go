package main

// Estructura auxiliar del Tracker para registrar los nodos (Peers)
type TorrentTrack struct {
	Peers map[string]*PeerInfo
}

// PeerInfo guarda los roles actuales (Leecher, Seeder), porcentaje de descarga y puertos asignados
type PeerInfo struct {
	IP             string
	Port           int
	Role           string  // "Leecher" o "Seeder"
	DownloadStatus float64 // Porcentaje de descarga
}
