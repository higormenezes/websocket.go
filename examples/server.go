package main

import (
	"fmt"
	"net/http"

	ws "github.com/higormenezes/websocket.go"
	"github.com/higormenezes/websocket.go/server"
)

func main() {

	wsServer := ws.Server{
		Config: ws.ServerConfig{
			Protocols: []string{"cursor", "chat", "test"},
		},
		HandleTextMessage: func(conn *server.WsConn, payload string) {
			fmt.Printf("%s: %s\n", conn.LocalAddr(), payload)
		},
	}

	// wsServer.HandleConnection = func(conn *server.WsConn) {
	// 	fmt.Printf("Connected %s\n", conn.LocalAddr())
	// }
	// wsServer.ConnKey
	// wsServer.HandleFunc
	// wsServer.OnDisconnect

	http.HandleFunc("/ws", wsServer.Handler)
	http.ListenAndServe(":8090", nil)
}
