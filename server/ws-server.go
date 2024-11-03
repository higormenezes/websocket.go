package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/higormenezes/websocket.go/internal"
)

type Config struct {
	Protocols []string
}

type Server struct {
	Config            Config
	HandleConnection  func(conn *WsConn)
	HandleDisconnect  func(conn *WsConn)
	HandleTextMessage func(conn *WsConn, payload string)
	HandleByteMessage func(conn *WsConn, payload []byte)
}

func (s Server) handshake(wsConn *WsConn, req *http.Request) error {
	if req.Method != http.MethodGet {
		return fmt.Errorf("invalid HTTP Method. To stablish a WebSocket connection, it is expected a %s Method", http.MethodGet)
	}

	webSocketKey := req.Header.Get("Sec-WebSocket-Key")
	acceptKey, err := internal.GetWebSocketAcceptKey(webSocketKey)
	if err != nil {
		return err
	}

	webSocketProtocol := req.Header.Get("Sec-WebSocket-Protocol")
	var lines []string
	lines = append(lines, "HTTP/1.1 101 Switching Protocols")
	lines = append(lines, "Upgrade: websocket")
	lines = append(lines, "Connection: Upgrade")
	lines = append(lines, "Sec-WebSocket-Accept: "+acceptKey)

	if webSocketProtocol != "" {
		protocols := strings.Split(webSocketProtocol, ",")
		protocolIndex := slices.IndexFunc(s.Config.Protocols, func(protocol string) bool {
			for _, headerProtocol := range protocols {
				if headerProtocol == protocol {
					return true
				}
			}
			return false
		})

		if protocolIndex != -1 {
			protocol := s.Config.Protocols[protocolIndex]
			lines = append(lines, "Sec-WebSocket-Protocol: "+protocol)
		}
	}
	lines = append(lines, "")
	lines = append(lines, "")

	_, err = wsConn.bufrw.WriteString(strings.Join(lines, "\r\n"))
	if err != nil {
		return err
	}

	err = wsConn.bufrw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (s Server) Handler(w http.ResponseWriter, req *http.Request) {
	wsConn := new(WsConn)

	err := wsConn.hijackHttpConnection(w)
	if err != nil {
		log.Println("hijackHttpConnection " + err.Error())
		return
	}
	defer wsConn.Close()

	err = s.handshake(wsConn, req)
	if err != nil {
		log.Println("handshake " + err.Error())
		return
	}

	if s.HandleConnection != nil {
		s.HandleConnection(wsConn)
	}

	var payloadData []byte
	var firstFrameOpcode uint8
	for {
		frame, err := wsConn.read()
		if err != nil {
			log.Println("GetFrameFromBuf " + err.Error())
			return
		}

		if len(payloadData) == 0 {
			firstFrameOpcode = frame.Opcode
		}
		payloadData = append(payloadData, frame.PayloadData...)

		// Control frames themselves MUST NOT be fragmented
		// fragment only allowed for non controlled messages

		opcode := frame.Opcode
		if frame.Fin && frame.Opcode == 0 {
			opcode = firstFrameOpcode
		}
		switch opcode {
		case 0: // continuation frame

		case 1: // text frame
			if frame.Fin {
				text := string(payloadData)
				if s.HandleTextMessage != nil {
					s.HandleTextMessage(wsConn, text)
				}
				payloadData = nil
			}

		case 2: // binary frame
			if frame.Fin {
				if s.HandleByteMessage != nil {
					s.HandleByteMessage(wsConn, payloadData)
				}
				payloadData = nil
			}

		case 3, 4, 5, 6, 7: // reserved for further non-control frames

		case 8: // connection close
			// TODO: Send close frame
			if s.HandleDisconnect != nil {
				s.HandleDisconnect(wsConn)
			}
			wsConn.Close()

		case 9: // ping
			frame.Opcode = 10
			wsConn.write(frame)

		case 10: // pong
		case 11, 12, 13, 14, 15: // reserved for further control frames
		default:
			log.Println("GetFrameFromBuf " + errors.New("no valid Opcode").Error())
			return
		}

	}

}
