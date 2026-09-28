# Cliente P2P Distribuido (Protocolo BitTorrent)

Implementación de un sistema distribuido de intercambio de archivos peer-to-peer (P2P) escrito en Go. Este proyecto replica la arquitectura central del protocolo BitTorrent, permitiendo la transferencia concurrente de archivos de gran tamaño a través de una red local (LAN) utilizando sockets TCP, fragmentación de datos y validación criptográfica.

## Características Principales
* **Descubrimiento Dinámico:** Tracker centralizado que mantiene el estado global de la red, registrando los puertos e IPs de los nodos activos y coordinando el enjambre (Swarm).
* **Descarga Concurrente:** Arquitectura multihilo que utiliza *Goroutines* y *Channels* para solicitar diferentes fragmentos de un mismo archivo a múltiples nodos de forma simultánea.
* **Tolerancia a Fallos:** Persistencia de estado local mediante archivos `.progress`. Ante una pérdida de conexión o cierre abrupto, el sistema retoma la descarga exactamente en el porcentaje donde se interrumpió.
* **Validación de Integridad:** Cálculo y verificación de hashes MD5 por cada fragmento recibido (tamaño de bloque fijo) para garantizar que los paquetes no sufran corrupción en la capa de red.
* **Soporte Multiplataforma:** Código agnóstico al sistema operativo, compilable a binarios nativos estáticos para Linux y Windows sin requerir dependencias de ejecución.

## Estructura de Directorios Requerida
Antes de ejecutar los nodos, el sistema requiere la siguiente estructura de carpetas en el mismo directorio donde se encuentre el archivo ejecutable:
* `archivos/`: Directorio donde se deben colocar los archivos originales que se desean compartir (Seeder), y donde el sistema guardará los ensamblajes finales (Leecher) utilizando el prefijo `descargado_`.
* `torrents/`: Directorio de metadatos de red. Aquí se generan y leen los archivos `.torrent` (mapa criptográfico) y `.progress` (registro json de pedazos completados).

## Compilación Cruzada
El sistema puede ser compilado desde cualquier entorno con Go 1.21+ hacia el sistema operativo de destino.

**Para despliegue nativo en Linux (Debian, Ubuntu, etc.):**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bittorrent_node .
```

**Para despliegue nativo en Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o bittorrent_node.exe .
```

## Instrucciones de Uso (Flujo de Red)

### 1. Iniciar el Nodo Coordinador (Tracker)
El Tracker es el núcleo de descubrimiento y debe iniciarse antes que cualquier cliente. Escuchará de forma predeterminada en el puerto `5000` aceptando conexiones de cualquier interfaz de red (`0.0.0.0`).
```bash
./bittorrent_node tracker
```
*Identifica la IP LAN de la máquina donde corre el Tracker (ej. `192.168.1.100`) para el siguiente paso.*

### 2. Generar Metadatos e Iniciar Compartición (Seeder)
Coloca el archivo objetivo dentro de la carpeta `archivos/`. Lanza un nodo Peer apuntando a la IP del Tracker.
```bash
./bittorrent_node peer 192.168.1.100
```
Dentro del menú interactivo de consola:
1. Selecciona la **Opción 1**.
2. Ingresa el nombre exacto de tu archivo (ej. `pelicula.mp4`).
3. El sistema segmentará el archivo binario, generará `pelicula.mp4.torrent` en la carpeta de metadatos y registrará tu IP/Puerto en el Tracker como el Seeder público inicial.

### 3. Descargar el Archivo (Leecher)
En el nodo destino, transfiere y coloca únicamente el archivo `.torrent` generado en el paso anterior dentro de su respectiva carpeta `torrents/`. Inicia el ejecutable:
```bash
./bittorrent_node peer 192.168.1.100
```
En el menú interactivo:
1. Selecciona la **Opción 1**.
2. Ingresa el nombre del archivo de metadatos (ej. `pelicula.mp4.torrent`).
3. La interfaz mostrará el descubrimiento de nodos disponibles e iniciará la transferencia P2P. Al llegar al 100%, el rol del nodo cambiará dinámicamente a Seeder para apoyar al enjambre.

## Autores
**Jesús Reynaldo Palma Hernández** 
**Karen Ayareth Ramírez Martínez**
**Hector Jair Mendoza Xicotencatl** 
Ingeniería Telemática, UPIITA-IPN