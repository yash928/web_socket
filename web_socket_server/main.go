package main

import (
	"fmt"
	"net/http"
	"web_socket_server/socket_server"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	server := socket_server.NewWebSocketServer()

	r.Get("/ws", server.HandleWebSocket)

	go server.StartBroadcast()

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
