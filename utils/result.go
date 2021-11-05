package utils

import (
	"bufio"
	"encoding/binary"
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

func GetBytes(reader *bufio.Reader, size uint64) ([]byte, error) {
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
	}
	_, err = reader.Discard(1)
	if err != nil {
		log.Println(err)
	}
	return state[0]
}

//Request build functions
func SetUint64(request *[]byte, data uint64) {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, data)
	*request = append(*request, bs...)
}
func SetBytes(request *[]byte, data []byte) {
	*request = append(*request, data...)
}
func SetByte(request *[]byte, data byte) {
	*request = append(*request, data)
}
