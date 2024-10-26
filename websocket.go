package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"log"
	"net/http"
	"strings"
)

// Client
//
// GET /chat HTTP/1.1
// Host: server.example.com
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
// Origin: http://example.com
// Sec-WebSocket-Protocol: chat, superchat
// Sec-WebSocket-Version: 13

// Sever
//
// HTTP/1.1 101 Switching Protocols
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
// Sec-WebSocket-Protocol: chat

const WebsocketHandshakeGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func getWebSocketAcceptKey(webSocketKey string) (string, error) {
	trimmedWebSocketKey := strings.Trim(webSocketKey, " ")
	hash := sha1.New()

	_, error := hash.Write([]byte(trimmedWebSocketKey + WebsocketHandshakeGUID))
	if error != nil {
		return "", error
	}

	acceptKey := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return acceptKey, nil
}

func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid web socket http method", http.StatusBadRequest)
		return
	}

	webSocketKey := req.Header.Get("Sec-WebSocket-Key")
	acceptKey, err := getWebSocketAcceptKey(webSocketKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Web server does not support hijacking", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	_, err = bufrw.WriteString(strings.Join([]string{
		"HTTP/1.1 101 Web Socket Protocol Handshake",
		"Server: go/echoserver",
		"Upgrade: WebSocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + acceptKey,
		"", // required for extra CRLF
		"", // required for extra CRLF
	}, "\r\n"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bufrw.Flush()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {

		buf := make([]byte, 2)
		_, err := io.ReadFull(bufrw, buf)
		if err != nil {
			log.Printf("Error 1")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		FIN := (buf[0] >> 7) & 1
		RSV1 := (buf[0] >> 6) & 1
		RSV2 := (buf[0] >> 5) & 1
		RSV3 := (buf[0] >> 4) & 1
		Opcode := buf[0] & 0b00001111

		log.Printf("FIN: %d; RSV1: %d; RSV2: %d; RSV3: %d; Opcode: %d", FIN, RSV1, RSV2, RSV3, Opcode)

		mask := (buf[1] >> 7) & 1
		payloadLength := uint64(buf[1] & 0b01111111)

		if payloadLength == 126 {
			lengthBuf := make([]byte, 2)
			_, err := io.ReadFull(bufrw, lengthBuf)
			if err != nil {
				log.Printf("Error 2")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			payloadLength = uint64(binary.BigEndian.Uint16(lengthBuf))
		} else if payloadLength == 127 {
			lengthBuf := make([]byte, 8)
			_, err := io.ReadFull(bufrw, lengthBuf)
			if err != nil {
				log.Printf("Error 3")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			payloadLength = uint64(binary.BigEndian.Uint16(lengthBuf))
		}

		log.Printf("mask: %d; payloadLength: %d", mask, payloadLength)

		maskKeyBuf := make([]byte, 4)
		if mask == 1 {
			_, err = io.ReadFull(bufrw, maskKeyBuf)
			if err != nil {
				log.Printf("Error 4")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		log.Printf("maskKeyBuf: %b", maskKeyBuf)

		payloadBuf := make([]byte, payloadLength)
		_, err = io.ReadFull(bufrw, payloadBuf)
		if err != nil {
			log.Printf("Error 5")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var payloadStrBuilder strings.Builder
		if mask == 1 {
			for idx, payloadOctet := range payloadBuf {
				err = payloadStrBuilder.WriteByte(payloadOctet ^ maskKeyBuf[idx%4])
				if err != nil {
					log.Printf("Error 6")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			_, err = payloadStrBuilder.Write(payloadBuf)
			if err != nil {
				log.Printf("Error 6")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		log.Printf("payload: %s", payloadStrBuilder.String())
	}

}
