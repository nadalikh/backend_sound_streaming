package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/go-mp3"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type msConnection struct {
	Ws   *websocket.Conn
	Id   uuid.UUID
	Type string
}
type channelData struct {
	Data []byte
	Id   uuid.UUID
	Type string
}

var connections = make(map[*msConnection]bool)
var data = make(chan channelData)
var mutex = sync.Mutex{}
var i int
var i1 int

func init() {
	i = 0
	i1 = 0
}

// var mutex = sync.Mutex{}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (adjust this based on your needs)
	},
}

// 6868518
func streamMusic(conn *msConnection) {

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
	data <- channelData{Data: []byte("seconds:" + strconv.FormatInt(soundDuration, 10)), Type: "stream", Id: conn.Id}
	data <- channelData{Data: []byte("size:" + strconv.FormatInt(soundSize, 10)), Type: "stream", Id: conn.Id}
	buf := make([]byte, soundSize) // 32KB buffer
	fmt.Println(bytesPerSecond)

	//for {
	i = i + 1

	//time.Sleep(1000 * time.Millisecond)

	_, err = file.Read(buf)

	if err == io.EOF {
		log.Println("Finished streaming")
		//break
	}
	if err != nil {
		log.Println("Error reading file:", err)
		//break
		//}
		//if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
		//	log.Println("Error writing message:", err)
		//	break
		//}
	}
	data <- channelData{Data: buf, Type: "stream", Id: conn.Id}

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	role := params.Get("role")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	msConn := msConnection{Ws: conn, Id: uuid.New(), Type: "stream"}

	mutex.Lock()
	connections[&msConn] = true
	mutex.Unlock()
	if role == "admin" {
		go streamMusic(&msConn)
	}
}
func broadcastMessage() {
	for {
		chunk := <-data
		i1 = i1 + 1

		mutex.Lock()
		//fmt.Println("sends data", chunk.Id)
		for conn := range connections {

			//fmt.Println(conn.Id != chunk.Id && conn.Type != "stream")
			if chunk.Type == "stream" {
				if conn.Id == chunk.Id {
					err := conn.Ws.WriteMessage(websocket.BinaryMessage, chunk.Data)
					if err != nil {
						log.Println("Error sending to client:", err)
						conn.Ws.Close()
						delete(connections, conn)
					}
				}
			} else if conn.Id != chunk.Id && conn.Type != "stream" {
				//fmt.Println("connection: ", conn.Id, conn.Type)
				//fmt.Println("data: ", chunk.Id, chunk.Type)
				err := conn.Ws.WriteMessage(websocket.BinaryMessage, chunk.Data)
				if err != nil {
					log.Println("Error sending to client:", err)
					conn.Ws.Close()
					delete(connections, conn)
				}
			}
		}
		mutex.Unlock()
	}
}
func stream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	msConn := msConnection{Ws: conn, Id: uuid.New()}

	mutex.Lock()
	connections[&msConn] = true
	mutex.Unlock()

	for {
		// Read message from the WebSocket connection

		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		// Print the received message to the console

		// Broadcast the received message to all connected clients
		data <- channelData{Data: p, Id: msConn.Id}

	}
}
func handleTimeWS(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new connection: ")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	msConn := msConnection{Ws: conn, Id: uuid.New()}
	mutex.Lock()
	connections[&msConn] = true
	mutex.Unlock()
	go func() {
		for {
			// Read message from the WebSocket connection
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}

			if len(p) > 0 {

				_, err := strconv.ParseFloat(string(p), 32)
				if err != nil {
					panic(err)
				}

				if err != nil {
					fmt.Println("Error reading message:", err)
					return
				}
				data <- channelData{Data: p, Id: msConn.Id}
			}
		}
	}()
}
func main() {
	http.HandleFunc("/ws/", handleWebSocket)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/time", handleTimeWS)
	go broadcastMessage()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe("172.26.200.194:8080", nil))
}
