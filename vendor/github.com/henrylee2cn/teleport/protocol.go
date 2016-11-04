package teleport

import (
	"bytes"
	"encoding/binary"
)

const (
	// 支持数据最大长度为 2 << 61
	// DataLengthOfLenth = 8
	// 支持数据最大长度为 2 << 30
	DataLengthOfLenth = 4
)

//通讯协议处理，主要处理封包和解包的过程
type Protocol struct {
	// 包头
	header string
	// 包头长度
	headerLen int
}

func NewProtocol(packetHeader string) *Protocol {
	return &Protocol{
		header:    packetHeader,
		headerLen: len([]byte(packetHeader)),
	}
}

func (self *Protocol) ReSet(header string) {
	self.header = header
	self.headerLen = len([]byte(header))
}

//封包
func (self *Protocol) Packet(message []byte) []byte {
	return append(append([]byte(self.header), IntToBytes(len(message))...), message...)
}

//解包
func (self *Protocol) Unpack(buffer []byte) (readerSlice [][]byte, bufferOver []byte) {
	length := len(buffer)

	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+self.headerLen+DataLengthOfLenth {
			break
		}
		if string(buffer[i:i+self.headerLen]) == self.header {
			messageLength := BytesToInt(buffer[i+self.headerLen : i+self.headerLen+DataLengthOfLenth])
			if length < i+self.headerLen+DataLengthOfLenth+messageLength {
				break
			}
			data := buffer[i+self.headerLen+DataLengthOfLenth : i+self.headerLen+DataLengthOfLenth+messageLength]

			readerSlice = append(readerSlice, data)

			i += self.headerLen + DataLengthOfLenth + messageLength - 1
		}
	}

	if i == length {
		bufferOver = make([]byte, 0)
		return
	}
	bufferOver = buffer[i:]
	return
}

//整形转换成字节
// func IntToBytes(n int) []byte {
// 	x := int64(n)

// 	bytesBuffer := bytes.NewBuffer([]byte{})
// 	binary.Write(bytesBuffer, binary.BigEndian, x)
// 	return bytesBuffer.Bytes()
// }

// //字节转换成整形
// func BytesToInt(b []byte) int {
// 	bytesBuffer := bytes.NewBuffer(b)

// 	var x int64
// 	binary.Read(bytesBuffer, binary.BigEndian, &x)

// 	return int(x)
// }

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)

	return int(x)
}
