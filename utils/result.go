package utils

import (
	"bufio"
	"encoding/binary"
	"github.com/DSN-team/core"
	"log"
)

//Result parse functions
func GetUint64Reader(clientReader *bufio.Reader) uint64 {
	state, _ := GetBytes(clientReader, 8)
	count := binary.BigEndian.Uint64(state)
	return count
}

func GetUint32Reader(clientReader *bufio.Reader) uint32 {
	state, _ := GetBytes(clientReader, 4)
	count := binary.BigEndian.Uint32(state)
	return count
}
func GetUint16Reader(clientReader *bufio.Reader) uint16 {
	state, _ := GetBytes(clientReader, 2)
	count := binary.BigEndian.Uint16(state)
	return count
}

func GetUint8Reader(clientReader *bufio.Reader) uint8 {
	state, _ := GetBytes(clientReader, 1)
	count := uint8(state[0])
	return count
}

func GetBytes(reader *bufio.Reader, size uint64) ([]byte, error) {
	if size == 0 {
		return make([]byte, 0), nil
	}
	state, err := reader.Peek(int(size))
	if err != nil {
		log.Println(err)
	}
	_, err = reader.Discard(int(size))
	if err != nil {
		log.Println(err)
	}
	return state, err
}
func GetByte(reader *bufio.Reader) byte {
	state, err := reader.Peek(1)
	if err != nil {
		log.Println(err)
		return core.RequestError
	}
	_, err = reader.Discard(1)
	if err != nil {
		log.Println(err)
		return core.RequestError
	}
	return state[0]
}

//Request build functions
func SetUint64(request *[]byte, data uint64) {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, data)
	*request = append(*request, bs...)
}
func SetUint32(request *[]byte, data uint32) {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, data)
	*request = append(*request, bs...)
}
func SetUint16(request *[]byte, data uint16) {
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, data)
	*request = append(*request, bs...)
}
func SetUint8(request *[]byte, data uint8) {
	*request = append(*request, data)
}

func SetBytes(request *[]byte, data []byte) {
	*request = append(*request, data...)
}
func SetSlice(request *[]byte, data []byte) {
	SetUint64(request, uint64(len(data)))
	*request = append(*request, data...)
}
func SetByte(request *[]byte, data byte) {
	*request = append(*request, data)
}
