package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type databaseStr struct {
	io []byte
}

var dataStr *databaseStr = &databaseStr{}
var dataStrInput *databaseStr = &databaseStr{}
var connClient net.Conn

func fakeClient() {
	for {
		time.Sleep(1 * time.Second)
		writeBytes([]byte{0, 1, 0})
	}
}
func main() {
	go analyzerRun()
	fakeClient()
}

//Инициализировать структуры и подключение
//export analyzerRun
func analyzerRun() {
	/*for {

		println("")
	}*/

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	connClient, err = net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
			// handle error
		}
		println("test")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	_, _ = conn.Read(dataStr.io)
	updateCall()
	time.Sleep(50)
}

//export writeBytes
func writeBytes(in []byte) {
	dataStrInput.io = in
	println("client:", connClient)
	_, _ = fmt.Fprint(connClient, dataStrInput.io)
}

//export exportBytes
func exportBytes() []byte {
	return dataStr.io
}

//Realisation for platform
func updateCall() {
	//Call Application to read structure and update internal data interpretations, update UI.

	//Test

	println((dataStr.io))
}
