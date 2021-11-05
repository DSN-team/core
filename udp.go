package core

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

type UdpStruct struct {
	con  net.Conn
	data []byte
}

func (u UdpStruct) sendData(profile *Profile, callback func()) {
	bs := make([]byte, 9)
	binary.BigEndian.PutUint64(bs, uint64(len(u.data)))
	bs[8] = '\n'
	bytes := append(bs, profile.DataStrInput.Io...)
	println("ClientSend:", bytes, " count:", len(u.data))

	if _, err = u.con.Write(bytes); err != nil {
		log.Printf("failed to send the client request: %v\n", err)
	}
}

func (u UdpStruct) startTimer() {
	go func() {
		for {
			doUdpStuff()
			time.Sleep(150)
		}
	}()
}

func doUdpStuff() {
	println("doing Udp stuff")
}
