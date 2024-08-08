package socket_server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketServer struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mu        sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

func (server *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	server.mu.Lock()
	fmt.Println("creating a new connection")
	server.clients[conn] = true
	server.mu.Unlock()

	go server.handleMessages(conn)
}

func (server *WebSocketServer) handleMessages(conn *websocket.Conn) {
	defer func() {
		server.mu.Lock()
		delete(server.clients, conn)
		server.mu.Unlock()
		conn.Close()
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		if messageType == websocket.TextMessage {
			server.broadcast <- p
		}
	}
}

func (server *WebSocketServer) StartBroadcast() {
	for {
		message := <-server.broadcast
		server.mu.Lock()
		for conn := range server.clients {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println("Write error:", err)
				conn.Close()
				delete(server.clients, conn)
			}
		}
		server.mu.Unlock()
	}
}
