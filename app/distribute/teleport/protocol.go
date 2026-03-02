package teleport

import (
	"bytes"
	"encoding/binary"
)

const (
	DataLengthOfLenth = 4
)

// Protocol handles packet framing (pack/unpack).
type Protocol struct {
	header    string
	headerLen int
}

func NewProtocol(packetHeader string) *Protocol {
	return &Protocol{
		header:    packetHeader,
		headerLen: len([]byte(packetHeader)),
	}
}

func (p *Protocol) ReSet(header string) {
	p.header = header
	p.headerLen = len([]byte(header))
}

// Packet frames a message for transmission.
func (p *Protocol) Packet(message []byte) []byte {
	return append(append([]byte(p.header), IntToBytes(len(message))...), message...)
}

// Unpack extracts messages from the buffer.
func (p *Protocol) Unpack(buffer []byte) (readerSlice [][]byte, bufferOver []byte) {
	length := len(buffer)

	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+p.headerLen+DataLengthOfLenth {
			break
		}
		if string(buffer[i:i+p.headerLen]) == p.header {
			messageLength := BytesToInt(buffer[i+p.headerLen : i+p.headerLen+DataLengthOfLenth])
			if length < i+p.headerLen+DataLengthOfLenth+messageLength {
				break
			}
			data := buffer[i+p.headerLen+DataLengthOfLenth : i+p.headerLen+DataLengthOfLenth+messageLength]

			readerSlice = append(readerSlice, data)

			i += p.headerLen + DataLengthOfLenth + messageLength - 1
		}
	}

	if i == length {
		bufferOver = make([]byte, 0)
		return
	}
	bufferOver = buffer[i:]
	return
}

// IntToBytes converts int to bytes.
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

// BytesToInt converts bytes to int.
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)

	return int(x)
}
