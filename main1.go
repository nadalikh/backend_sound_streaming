package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/go-mp3"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

var connections = make(map[*websocket.Conn]bool)
var data = make(chan []byte)
var mutex = sync.Mutex{}
var i int
var i1 int

func init() {
	i = 0
	i1 = 0
}

// var mutex = sync.Mutex{}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (adjust this based on your needs)
	},
}

// 6868518
func streamMusic(conn *websocket.Conn) {
	fmt.Println("herer")
	file, err := os.Open("sample.mp3") // Ensure your path is correct
	if err != nil {
		panic(err)
	}
	status, err := file.Stat()
	if err != nil {
		panic(err)
	}
	soundSize := status.Size()
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		log.Fatalf("Error decoding MP3: %v", err)
	}

	// Get the total duration in seconds
	soundDuration := int64(math.Floor(float64(decoder.Length()) / float64(decoder.SampleRate()) / 4))
	bytesPerSecond := soundSize / soundDuration
	fmt.Println(bytesPerSecond)

	buf := make([]byte, bytesPerSecond) // 32KB buffer
	for {
		i = i + 1

		time.Sleep(1000 * time.Millisecond)

		n, err := file.Read(buf)

		if err == io.EOF {
			log.Println("Finished streaming")
			break
		}
		if err != nil {
			log.Println("Error reading file:", err)
			break
		}
		//if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
		//	log.Println("Error writing message:", err)
		//	break
		//}
		data <- buf[:n]
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleWebSocket")
	params := r.URL.Query()
	role := params.Get("role")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	mutex.Lock()
	connections[conn] = true
	mutex.Unlock()
	if role == "admin" {
		go streamMusic(conn)
	}
}
func broadcastMessage() {
	for {
		chunk := <-data
		i1 = i1 + 1

		mutex.Lock()

		for conn := range connections {
			err := conn.WriteMessage(websocket.BinaryMessage, chunk)
			if err != nil {
				log.Println("Error sending to client:", err)
				conn.Close()
				delete(connections, conn)
			}
		}
		mutex.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws/", handleWebSocket)
	go broadcastMessage()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
