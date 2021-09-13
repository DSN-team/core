package main

import (
	"bufio"
	"encoding/binary"
	"github.com/ClarkGuan/jni"
	"io"
	"log"
	"net"
	"reflect"
	"runtime"
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
var connections = make(map[string]net.Conn)

var callBackBufferPtr unsafe.Pointer
var callBackBufferCap int
var workingVM jni.VM

func main() {
	println("main started")
}

//export Java_com_dsnteam_dsn_CoreManager_runClient
func Java_com_dsnteam_dsn_CoreManager_runClient(env uintptr, _ uintptr, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))
	println("env run client:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}
	go handleClientConnect(address)
}

func handleClientConnect(address string) {
	con, _ := net.Dial("tcp", address)
	connections[con.RemoteAddr().String()] = con
}

func server(address string) {
	ln, err := net.Listen("tcp", address)
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(ln)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		con, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handleServerConnection(con)
	}
}

//Инициализировать структуры и подключение

//export Java_com_dsnteam_dsn_CoreManager_runServer
func Java_com_dsnteam_dsn_CoreManager_runServer(env uintptr, _ uintptr, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))
	println("env run:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}

	go server(address)
}

//export Java_com_dsnteam_dsn_CoreManager_writeCallBackBuffer
func Java_com_dsnteam_dsn_CoreManager_writeCallBackBuffer(env uintptr, _ uintptr, jniBuffer uintptr) {
	callBackBufferPtr = jni.Env(env).GetDirectBufferAddress(jniBuffer)
	callBackBufferCap = jni.Env(env).GetDirectBufferCapacity(jniBuffer)
}

func handleServerConnection(con net.Conn) {
	defer func(con net.Conn) {
		err := con.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(con)
	println("handling")

	clientReader := bufio.NewReader(con)
	clientWriter := bufio.NewWriter(con)
	connections[con.RemoteAddr().String()] = con
	println(con.RemoteAddr().String())
	println("bufio")
	for {
		// Waiting for the client request
		println("reading")
		var err error
		state, err := clientReader.Peek(9)
		_, err = clientReader.Discard(9)
		count := binary.BigEndian.Uint64(state[0:8])
		println("Count:", count)
		dataStr.io, err = clientReader.Peek(int(count))
		_, err = clientReader.Discard(int(count))
		switch err {
		case nil:
			log.Println(dataStr.io)
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}
		// Responding to the client request
		updateCall(int(count))
		if _, err = clientWriter.Write([]byte("Accepted\n")); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}
	}
}

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, _ uintptr, inBuffer uintptr, lenIn int, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))

	println(address)
	for k := range connections {
		println(k)
	}

	println("env write:", env)
	defer runtime.KeepAlive(dataStrInput.io)
	point := jni.Env(env).GetDirectBufferAddress(inBuffer)
	size := jni.Env(env).GetDirectBufferCapacity(inBuffer)

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dataStrInput.io))
	sh.Data = uintptr(point)
	sh.Len = lenIn
	sh.Cap = size

	runtime.KeepAlive(dataStrInput.io)
	log.Println("input:", dataStrInput.io)
	println("input str:", string(dataStrInput.io))

	var err error
	switch err {
	case nil:
		bs := make([]byte, 9)
		binary.BigEndian.PutUint64(bs, uint64(lenIn))
		bs[8] = '\n'
		bytes := append(bs, dataStrInput.io...)
		println("ClientSend:", bytes, " count:", lenIn)
		if _, err = connections[address].Write(bytes); err != nil {
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
	var serverResponse []byte
	serverReader := bufio.NewReader(connections[address])
	serverResponse, err = serverReader.ReadBytes('\n')

	switch err {
	case nil:
		log.Println(serverResponse)
	case io.EOF:
		log.Println("server closed the connection")
		return
	default:
		log.Printf("server error: %v\n", err)
		return
	}
}

//export Java_com_dsnteam_dsn_CoreManager_exportBytes
func Java_com_dsnteam_dsn_CoreManager_exportBytes(env uintptr, clazz uintptr) uintptr {
	println("env export:", env)
	buffer := jni.Env(env).NewDirectByteBuffer(unsafe.Pointer(&dataStr.io), len(dataStr.io))
	return buffer
}

//Realisation for platform
func updateCall(count int) {
	//Call Application to read structure and update internal data interpretations, update UI.
	var env jni.Env
	env, _ = workingVM.AttachCurrentThread()
	//Test
	println(dataStr.io)
	println("WorkingEnv:", env)
	classInput := env.FindClass("com/dsnteam/dsn/CoreManager")
	println("class_input:", classInput)
	methodId := env.GetStaticMethodID(classInput, "getUpdateCallBack", "(I)V")
	println("MethodID:", methodId)
	var bData []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&bData))
	sh.Data = uintptr(callBackBufferPtr)
	sh.Cap = callBackBufferCap
	sh.Len = len(dataStr.io)
	println("buffer pointer:", callBackBufferPtr)
	copy(bData, dataStr.io)
	println("buffer write done")
	env.CallStaticVoidMethodA(classInput, methodId, jni.Jvalue(count))
	workingVM.DetachCurrentThread()
	runtime.KeepAlive(bData)
}
