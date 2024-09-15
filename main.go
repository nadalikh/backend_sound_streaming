package main

//
//import (
//	"log"
//	"net/http"
//	"time"
//
//	"github.com/gorilla/websocket"
//)
//
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//func handleWebSocket(w http.ResponseWriter, r *http.Request) {
//	// Upgrade HTTP connection to WebSocket
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println("Error upgrading to WebSocket:", err)
//		return
//	}
//	defer conn.Close()
//
//	// Set up ping-pong to keep the connection alive
//	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
//	conn.SetPongHandler(func(appData string) error {
//		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
//		return nil
//	})
//
//	// Start sending periodic pings to keep connection alive
//	ticker := time.NewTicker(30 * time.Second)
//	defer ticker.Stop()
//
//	go func() {
//		for {
//			<-ticker.C
//			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//				log.Println("Error sending ping:", err)
//				return
//			}
//		}
//	}()
//
//	// Send a simple message to the client
//	message := "Hello from the WebSocket server!"
//	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
//	if err != nil {
//		log.Println("Error writing message:", err)
//		return
//	}
//
//	// Keep the connection open to read messages from the client
//	for {
//		_, _, err := conn.ReadMessage()
//		if err != nil {
//			log.Println("Error reading message:", err)
//			break
//		}
//	}
//}
//
//func main() {
//	http.HandleFunc("/ws", handleWebSocket)
//	log.Println("WebSocket server started on :8080")
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}
