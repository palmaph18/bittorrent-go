package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . <tracker|peer> [IP_DEL_TRACKER]")
		return
	}

	mode := os.Args[1]

	if mode == "tracker" {
		// 0.0.0.0 permite que el Tracker reciba conexiones desde otras computadoras en la red LAN
		fmt.Println("[*] Iniciando Tracker en 0.0.0.0:5000...")
		tracker := NewTracker()
		tracker.Start("0.0.0.0:5000") // Asegúrate de que este método coincida con el que ya tienes
	} else if mode == "peer" {
		trackerIP := "127.0.0.1" // Fallback local si no se pasa el argumento

		// Leer la IP de la otra computadora si se proporcionó en la terminal
		if len(os.Args) >= 3 {
			trackerIP = os.Args[2]
		}

		trackerAddr := fmt.Sprintf("%s:5000", trackerIP)
		fmt.Printf("[*] Iniciando Peer... Apuntando al Tracker en %s\n", trackerAddr)

		peer := NewPeer(trackerAddr)
		peer.Start() // Asegúrate de que este método coincida con el que ya tienes
	} else {
		fmt.Println("[!] Modo no reconocido. Usa 'tracker' o 'peer'.")
	}
}
