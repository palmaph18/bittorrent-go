package main

import (
	"fmt"
	"os"
)

func main() {
	// Verificamos que se haya pasado al menos un argumento
	if len(os.Args) < 2 {
		fmt.Println("Uso incorrecto. Debes especificar el modo de ejecución.")
		fmt.Println("Ejemplo: go run . tracker")
		fmt.Println("Ejemplo: go run . peer")
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "tracker":
		// Iniciar el servidor centralizado (Capa de Monitoreo)
		tracker := NewTracker()
		tracker.Start()

	case "peer":
		// Iniciar un nodo actuando simultáneamente como cliente y servidor
		// Como todo será local, apuntamos al puerto 5000 por defecto del Tracker
		trackerAddr := "127.0.0.1:5000"
		fmt.Printf("Iniciando nodo... Conectando al Tracker en %s\n", trackerAddr)

		peer := NewPeer(trackerAddr)
		peer.Start()

	default:
		fmt.Printf("Modo '%s' no reconocido. Usa 'tracker' o 'peer'.\n", mode)
	}
}
