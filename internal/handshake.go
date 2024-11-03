package internal

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

const WebsocketHandshakeGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func GetWebSocketAcceptKey(webSocketKey string) (string, error) {
	trimmedWebSocketKey := strings.Trim(webSocketKey, " ")
	hash := sha1.New()

	_, error := hash.Write([]byte(trimmedWebSocketKey + WebsocketHandshakeGUID))
	if error != nil {
		return "", error
	}

	acceptKey := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return acceptKey, nil
}
