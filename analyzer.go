package main

import (
	"bufio"
	"fmt"
	"github.com/ClarkGuan/jni"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"strings"
	"unsafe"
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
var serverReader *bufio.Reader
var callBackBufferPtr unsafe.Pointer
var callBackBufferCap int
var workingVM jni.VM
var address string

func fakeClient() {
	/*for {
		time.Sleep(1 * time.Second)
		Java_com_dsnteam_dsn_CoreManager_writeBytes([]byte("testing\n"))
	}*/
	println("dial address:", address)
	connClient, _ = net.Dial("tcp", address)

	clientReader := bufio.NewReader(os.Stdin)

	serverReader = bufio.NewReader(connClient)
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
	address = ":8080"
	go analyzer()
	go fakeClient()
}

//export Java_com_dsnteam_dsn_CoreManager_connectionTarget
func Java_com_dsnteam_dsn_CoreManager_connectionTarget(env uintptr, clazz uintptr, stringIn uintptr) {
	address = string(jni.Env(env).GetStringUTF(stringIn))
}

//export Java_com_dsnteam_dsn_CoreManager_runClient
func Java_com_dsnteam_dsn_CoreManager_runClient(env uintptr, clazz uintptr) {
	println("envrunclient:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}
	go func() {
		connClient, _ = net.Dial("tcp", address)
		serverReader = bufio.NewReader(connClient)

	}()
	//go fakeClient()
}
func analyzer() {
	ln, err := net.Listen("tcp", address)
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

//Инициализировать структуры и подключение

//export Java_com_dsnteam_dsn_CoreManager_runAnalyzer
func Java_com_dsnteam_dsn_CoreManager_runAnalyzer(env uintptr, clazz uintptr) {
	println("envrun:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}

	go analyzer()
}

//export Java_com_dsnteam_dsn_CoreManager_writeCallBackBuffer
func Java_com_dsnteam_dsn_CoreManager_writeCallBackBuffer(env uintptr, clazz uintptr, jbuffer uintptr) {
	//callBackBuffer = jbuffer
	callBackBufferPtr = jni.Env(env).GetDirectBufferAddress(jbuffer)
	callBackBufferCap = jni.Env(env).GetDirectBufferCapacity(jbuffer)
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

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, clazz uintptr, inBuffer uintptr) {
	println("envwrite:", env)
	point := jni.Env(env).GetDirectBufferAddress(inBuffer)
	size := jni.Env(env).GetDirectBufferCapacity(inBuffer)

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dataStrInput.io))
	sh.Data = uintptr(point)
	sh.Len = size
	sh.Cap = size
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = dataStrInput.io[i]
	}
	runtime.KeepAlive(dataStrInput.io)
	println("inputstr:", string(dataStrInput.io))
	clientRequest := string(dataStrInput.io)
	var err error
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
	_, _ = fmt.Fprint(connClient, dataStrInput.io)
}

//export Java_com_dsnteam_dsn_CoreManager_exportBytes
func Java_com_dsnteam_dsn_CoreManager_exportBytes(env uintptr, clazz uintptr) uintptr {
	println("envexport:", env)
	buffer := jni.Env(env).NewDirectByteBuffer(unsafe.Pointer(&dataStr.io), len(dataStr.io))
	return buffer
}

//Realisation for platform
func updateCall() {
	//Call Application to read structure and update internal data interpretations, update UI.
	var curenv jni.Env
	curenv, _ = workingVM.AttachCurrentThread()
	//Test
	println((dataStr.io))
	println("WorkingEnv:", curenv)
	classinput := curenv.FindClass("com/dsnteam/dsn/CoreManager")
	println("classinput:", classinput)
	methodid := curenv.GetStaticMethodID(classinput, "getUpdateCallBack", "()V")
	println("MethodID:", methodid)
	var bdata []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&bdata))
	sh.Data = uintptr(callBackBufferPtr)
	sh.Len = callBackBufferCap
	sh.Cap = callBackBufferCap

	println("buffer pointer:", callBackBufferPtr)
	for i := 0; i < len(bdata); i++ {
		if i < len(dataStr.io) {
			bdata[i] = dataStr.io[i]
		}
	}
	println("buffer write done")
	curenv.CallStaticVoidMethodA(classinput, methodid)
	workingVM.DetachCurrentThread()
	runtime.KeepAlive(bdata)
}
