# Proyecto final de sistemas distribuidos: BitTorrent

## Descripción del proyecto
Es una red P2P tipo BitTorrent desarrollada en Go con CLI. El proyecto implementa una arquitectura descentralizada para la transferencia simultánea de al menos 6 archivos, con un tamaño mínimo de 200 MB cada uno. La implementación debe garantizar transparencia, concurrencia, tolerancia a fallos y recuperación de estado en caso de desconexiones.

## Rol
Eres un ingeniero de software senior con experiencia programando en Go, usando Debian con WSL2 y en proyectos de sistemas distribuidos quien siempre sigue buenas prácticas de programación.

## Tarea
Me vas a ayudar a diseñar y desarrollar una red BitTorrent en Go, asegurando que se cumplan las métricas de evaluación de sistemas distribuidos.

## Stack tecnológico
* Lenguaje de programación: Go (utilizando `goroutines` y `channels` para la concurrencia)
* Entorno: WSL2 con Debian 13
* IDE: Visual Studio Code
* Todo será local, es decir, nuevas terminales serán los nodos.

## Arquitectura y Componentes
Los módulos/paquetes principales a desarrollar para el funcionamiento del protocolo son:

*   **`torrent.go` (Capa de Datos):** Manejo de metadatos JSON (usando `encoding/json`), cálculo de hashes MD5 para validación de pedazos (tamaño sugerido de pedazo: 102400 bytes), y binarización.
*   **`tracker.go` (Capa de Monitoreo):** Servidor centralizado (HTTP o Sockets TCP) que gestiona solicitudes de conexión, administra los puertos de la red y mantiene un mapa (`map[string]*TorrentTrack`) con el estado global.
*   **`torrent_track.go`:** Estructura auxiliar del Tracker para registrar los nodos (Peers), sus roles actuales (Leecher, Seeder), porcentaje de descarga y puertos asignados.
*   **`peer.go` (Capa de Transferencia):** Lógica principal del nodo. Actúa simultáneamente como cliente y servidor lanzando múltiples *goroutines* para gestionar cargas y descargas concurrentes de pedazos.
*   **`connection.go`:** Maneja la información, persistencia y estado de cada socket TCP activo (conexión) con otros peers.
*   **`message.go`:** Define la estructura de los mensajes de la red y las constantes de tipos de mensaje (ej. ANNOUNCE, REQUEST_PIECE). Se encargará de la serialización a bytes (usando `encoding/gob` o `encoding/binary`).

## Políticas y Reglas del Sistema
*   **Regla de distribución:** Un nodo comenzará a compartir los fragmentos de un archivo hacia la red automáticamente una vez que haya superado el 20% de su descarga.
*   **Transferencia Simultánea:** Un mismo nodo debe ser capaz de solicitar y descargar piezas de un archivo desde múltiples nodos al mismo tiempo.
*   **Roles Dinámicos:** Los nodos inician como *Leechers* (si solo descargan) o *Seeders* (si poseen el archivo completo) y cambian dinámicamente según su estado.

## Condiciones en el desarrollo
*   **Tolerancia a fallos (Recuperación):** Si un nodo se desconecta, al volver a conectarse debe retomar las descargas y cargas exactamente en el porcentaje de progreso en el que se quedó.
*   El tracker preguntará periódicamente el estado de los nodos.
*   El Tracker operará por defecto en el puerto 5000.
*   Los puertos de los nodos (Peer) irán de 1000 en 1000 comenzando a partir del 4000 (ej. 4000, 6000, 7000) para evitar colisiones.
*   Definir un mecanismo para limpiar puertos y liberar (dejar ir) a un nodo cuando el Tracker detecte que se ha desconectado.
*   *Restricción:* No se incluirá criptografía en esta versión (se omite Diffie-Hellman).

## Estructura de los directorios
*   `/torrents`: Donde se guardan los archivos `.torrent` (metadatos en JSON) generados.
*   `/archivos`: Conjunto de archivos fuente a transferir (mínimo 6 archivos de >200MB).

## Menú CLI en cada nodo
Cada nodo (y el tracker) tendrá un menú CLI interactivo o un panel de registro (logs) que cumpla con los requerimientos visuales:
1. Agregar un torrent (iniciar descarga o compartir).
2. Debuggeo (mostrar logs de conexión).
3. Ver progreso local de las descargas (porcentaje o barra de progreso).
4. **Ver estado global de la red:** Lista de nodos conectados, roles (Leecher/Seeder), archivos compartidos, archivos en descarga y progresos (Esta opción es obligatoria para el Tracker).

## Para iniciar
1. Ayudame a configurar Visual Studio Code para programar en Go como los archivos iniciales y si es necesario, extensiones. Además de comprobar que tengo Go.
2. Dame la estructura de los directorios y archivos para crearla en la terminal de WSL2, recuerda tambien incluir el archivo `instrucciones.md`.
3. Después de completar los pasos anteriores, ahora si, a programar.

## Mejoras para el futuro
* Los nodos serán máquinas físicas o máquinas virtuales. Pero primero que funcione correctamente en local.