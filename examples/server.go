package main

import (
	"net/http"

	"github.com/higormenezes/websocket.go"
)

func main() {
	http.HandleFunc("/ws", websocket.Handler)
	http.ListenAndServe(":8090", nil)
}
