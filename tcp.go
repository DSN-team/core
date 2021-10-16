package main

import (
	"encoding/binary"
	"log"
	"net"
)

type TcpStruct struct {
	con  net.Conn
	data []byte
}

func (u TcpStruct) sendData(callback func()) {
	bs := make([]byte, 9)
	binary.BigEndian.PutUint64(bs, uint64(len(u.data)))
	bs[8] = '\n'
	bytes := append(bs, dataStrInput.io...)
	println("ClientSend:", bytes, " count:", len(u.data))

	if _, err = u.con.Write(bytes); err != nil {
		log.Printf("failed to send the client request: %v\n", err)
	}
}

func (u TcpStruct) startTimer() {
	println("There's no timer fo TCP")
}
