package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
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

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		i = i + 1
		fmt.Println(i)
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
		fmt.Println("i2: ", i1)
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
