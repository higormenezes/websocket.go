package internal

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const websocketHandshakeGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func getWebSocketAcceptKey(webSocketKey string) (string, error) {
	trimmedWebSocketKey := strings.Trim(webSocketKey, " ")
	hash := sha1.New()

	_, error := hash.Write([]byte(trimmedWebSocketKey + websocketHandshakeGUID))
	if error != nil {
		return "", error
	}

	acceptKey := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return acceptKey, nil
}

func Handshake(w http.ResponseWriter, req *http.Request) (net.Conn, *bufio.ReadWriter, error) {
	var conn net.Conn
	var bufioReadWriter *bufio.ReadWriter

	if req.Method != http.MethodGet {
		return conn, bufioReadWriter, fmt.Errorf("invalid HTTP Method. To stablish a WebSocket connection, it is expected a %s Method", http.MethodGet)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Web server does not support hijacking", http.StatusInternalServerError)
		return conn, bufioReadWriter, errors.New("web server does not support hijacking")
	}

	conn, bufioReadWriter, err := hj.Hijack()
	if err != nil {
		return conn, bufioReadWriter, err
	}

	webSocketKey := req.Header.Get("Sec-WebSocket-Key")
	acceptKey, err := getWebSocketAcceptKey(webSocketKey)
	if err != nil {
		return conn, bufioReadWriter, err
	}

	_, err = bufioReadWriter.WriteString(strings.Join([]string{
		"HTTP/1.1 101 Switching Protocols",
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + acceptKey,
		// "Sec-WebSocket-Protocol: chat", // TODO: SubProtocols
		"",
		"",
	}, "\r\n"))
	if err != nil {
		return conn, bufioReadWriter, err
	}

	err = bufioReadWriter.Flush()
	if err != nil {
		return conn, bufioReadWriter, err
	}

	return conn, bufioReadWriter, nil
}
