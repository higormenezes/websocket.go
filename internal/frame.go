package internal

type Frame struct {
	Fin, Rsv1, Rsv2, Rsv3 bool
	Opcode                uint8
	Masked                bool
	PayloadLength         uint64
	MaskingKey            [4]byte
	PayloadData           []byte
}
