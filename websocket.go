package websocket

import (
	"log"
	"net/http"

	"github.com/higormenezes/websocket.go/internal"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	conn, bufioReadWriter, err := internal.Handshake(w, req)
	if err != nil {
		log.Println("Handshake " + err.Error())
		return
	}
	defer conn.Close()

	for {
		_, err := internal.GetFrameFromBuf(bufioReadWriter)
		if err != nil {
			log.Println("GetFrameFromBuf " + err.Error())
			return
		}

		// log.Printf("%+v\n", frame)
	}

}
