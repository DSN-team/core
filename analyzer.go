package main

import (
	"bufio"
	"fmt"
	"github.com/ClarkGuan/jni"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// #include <stdlib.h>
// #include <stddef.h>
// #include <stdint.h>
import "C"

type databaseStr struct {
	io []byte
}

var dataStr = &databaseStr{}
var dataStrInput = &databaseStr{}
var connClient net.Conn
var workingEnv jni.Env
var workingClazz jni.Jclass

func fakeClient() {
	/*for {
		time.Sleep(1 * time.Second)
		writeBytes([]byte("testing\n"))
	}*/
	connClient, _ = net.Dial("tcp", ":8080")

	clientReader := bufio.NewReader(os.Stdin)

	serverReader := bufio.NewReader(connClient)
	for {
		// Waiting for the client request
		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if _, err = connClient.Write([]byte(clientRequest + "\n")); err != nil {
				log.Printf("failed to send the client request: %v\n", err)
			}
		case io.EOF:
			log.Println("client closed the connection")
			return
		default:
			log.Printf("client error: %v\n", err)
			return
		}

		// Waiting for the server response
		serverResponse, err := serverReader.ReadString('\n')

		switch err {
		case nil:
			log.Println(strings.TrimSpace(serverResponse))
		case io.EOF:
			log.Println("server closed the connection")
			return
		default:
			log.Printf("server error: %v\n", err)
			return
		}
	}
}
func main() {
	go jni_com_dsnteam_runanalyzer(0, 0)
	fakeClient()
}

//Инициализировать структуры и подключение
//export jni_com_dsnteam_runanalyzer
func jni_com_dsnteam_runanalyzer(env uintptr, clazz uintptr) {
	if env != 0 {
		workingEnv = jni.Env(env)
	}
	if clazz != 0 {
		workingClazz = clazz
	}
	/*for {

		println("")
	}*/
	ln, err := net.Listen("tcp", ":8080")
	defer ln.Close()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(con net.Conn) {
	defer con.Close()
	println("handling")
	clientReader := bufio.NewReader(con)
	println("bufio")
	for {
		// Waiting for the client request
		println("reading")
		clientRequest, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == ":QUIT" {
				log.Println("client requested server to close the connection so closing")
				return
			} else {
				log.Println(clientRequest)
			}
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}

		// Responding to the client request
		dataStr.io = []byte(clientRequest)
		updateCall()
		if _, err = con.Write([]byte("GOT IT!\n")); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}
	}

}

//export writeBytes
func writeBytes(in []byte) {
	dataStrInput.io = in
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

	//methodid := workingEnv.GetStaticMethodID(workingClazz,"getUpdateCallBack","(L)")
	//workingEnv.CallStaticObjectMethodA(workingClazz,methodid,)
}
