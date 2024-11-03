package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/higormenezes/websocket.go/internal"
)

type WsConn struct {
	conn  net.Conn
	bufrw *bufio.ReadWriter
}

func (c *WsConn) hijackHttpConnection(w http.ResponseWriter) error {
	conn, bufrw, err := internal.HijackHttpConnection(w)
	if err != nil {
		return err
	}
	c.conn = conn
	c.bufrw = bufrw

	return nil
}

func (c *WsConn) Close() error {
	return c.conn.Close()
}

func (c *WsConn) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

func (c *WsConn) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *WsConn) read() (internal.Frame, error) {
	var frame internal.Frame

	tempBuf := make([]byte, 2)
	_, err := io.ReadFull(c.bufrw, tempBuf)
	if err != nil {
		log.Println("Error 1")
		return frame, err
	}

	frame.Fin = (tempBuf[0] >> 7) == 1
	frame.Rsv1 = (tempBuf[0] >> 6) == 1
	frame.Rsv2 = (tempBuf[0] >> 5) == 1
	frame.Rsv3 = (tempBuf[0] >> 4) == 1
	frame.Opcode = uint8((tempBuf[0] & 0b00001111))

	frame.Masked = (tempBuf[1] >> 7) == 1
	tempLength := uint8(tempBuf[1] & 0b01111111)

	if tempLength == 126 {
		lengthBuf := make([]byte, 2)
		_, err := io.ReadFull(c.bufrw, lengthBuf)
		if err != nil {
			return frame, err
		}
		frame.PayloadLength = uint64(binary.BigEndian.Uint16(lengthBuf))
	} else if tempLength == 127 {
		lengthBuf := make([]byte, 8)
		_, err := io.ReadFull(c.bufrw, lengthBuf)
		if err != nil {
			return frame, err
		}
		frame.PayloadLength = uint64(binary.BigEndian.Uint64(lengthBuf))
	} else {
		frame.PayloadLength = uint64(tempLength)
	}

	if frame.Masked {
		_, err = io.ReadFull(c.bufrw, frame.MaskingKey[:])
		if err != nil {
			return frame, err
		}
	}

	payloadBuf := make([]byte, frame.PayloadLength)
	_, err = io.ReadFull(c.bufrw, payloadBuf)
	if err != nil {
		return frame, err
	}

	var payloadStrBuilder bytes.Buffer
	if frame.Masked {
		for idx, payloadByte := range payloadBuf {
			err = payloadStrBuilder.WriteByte(payloadByte ^ frame.MaskingKey[idx%4])
			if err != nil {
				return frame, err
			}
		}
	} else {
		_, err = payloadStrBuilder.Write(payloadBuf)
		if err != nil {
			return frame, err
		}
	}
	frame.PayloadData = payloadStrBuilder.Bytes()

	return frame, nil
}

func (c *WsConn) write(frame internal.Frame) error {
	var bytesBuffer bytes.Buffer

	var tempByte byte
	if frame.Fin {
		tempByte |= 0b10000000
	}
	if frame.Rsv1 {
		tempByte |= 0b01000000
	}
	if frame.Rsv2 {
		tempByte |= 0b00100000
	}
	if frame.Rsv3 {
		tempByte |= 0b00010000
	}
	tempByte |= frame.Opcode

	err := bytesBuffer.WriteByte(tempByte)
	if err != nil {
		return err
	}
	tempByte = 0

	if frame.Masked {
		tempByte |= 0b10000000
	}

	if frame.PayloadLength < 126 {
		tempByte |= uint8(frame.PayloadLength)

		err = bytesBuffer.WriteByte(tempByte)
		if err != nil {
			return err
		}
		tempByte = 0

	} else if frame.PayloadLength == uint64(uint16(frame.PayloadLength)) {
		tempByte |= 126

		err = bytesBuffer.WriteByte(tempByte)
		if err != nil {
			return err
		}
		tempByte = 0

		lengthBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lengthBuf, uint16(frame.PayloadLength))
		_, err = bytesBuffer.Write(lengthBuf)
		if err != nil {
			return err
		}

	} else {
		tempByte |= 127

		err = bytesBuffer.WriteByte(tempByte)
		if err != nil {
			return err
		}
		tempByte = 0

		lengthBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(lengthBuf, uint64(frame.PayloadLength))
		_, err = bytesBuffer.Write(lengthBuf)
		if err != nil {
			return err
		}
	}

	if frame.Masked {
		_, err = bytesBuffer.Write(frame.MaskingKey[:])
		if err != nil {
			return err
		}
	}

	if frame.Masked {
		for idx, payloadDataByte := range frame.PayloadData {
			err = bytesBuffer.WriteByte(payloadDataByte ^ frame.MaskingKey[idx%4])
			if err != nil {
				return err
			}
		}
	} else {
		_, err = bytesBuffer.Write(frame.PayloadData)
		if err != nil {
			return err
		}
	}

	_, err = c.bufrw.Write(bytesBuffer.Bytes())
	if err != nil {
		return err
	}
	err = c.bufrw.Flush()
	if err != nil {
		return err
	}

	return nil
}
