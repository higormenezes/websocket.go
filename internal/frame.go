package internal

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"strings"
)

type Frame struct {
	Fin, Rsv1, Rsv2, Rsv3 bool
	Opcode                uint8
	Masked                bool
	PayloadLength         uint64
	MaskingKey            [4]byte
	PayloadData           string
}

func GetFrameFromBuf(bufioReadWriter *bufio.ReadWriter) (Frame, error) {
	var frame Frame

	tempBuf := make([]byte, 2)
	_, err := io.ReadFull(bufioReadWriter, tempBuf)
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
		_, err := io.ReadFull(bufioReadWriter, lengthBuf)
		if err != nil {
			return frame, err
		}
		frame.PayloadLength = uint64(binary.BigEndian.Uint16(lengthBuf))
	} else if tempLength == 127 {
		lengthBuf := make([]byte, 8)
		_, err := io.ReadFull(bufioReadWriter, lengthBuf)
		if err != nil {
			return frame, err
		}
		frame.PayloadLength = uint64(binary.BigEndian.Uint64(lengthBuf))
	} else {
		frame.PayloadLength = uint64(tempLength)
	}

	if frame.Masked {
		_, err = io.ReadFull(bufioReadWriter, frame.MaskingKey[:])
		if err != nil {
			return frame, err
		}
	}

	payloadBuf := make([]byte, frame.PayloadLength)
	_, err = io.ReadFull(bufioReadWriter, payloadBuf)
	if err != nil {
		return frame, err
	}

	var payloadStrBuilder strings.Builder
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
	frame.PayloadData = payloadStrBuilder.String()

	log.Printf("%+v\n", frame)

	return frame, nil
}
