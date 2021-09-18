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
	"sync"
	"unsafe"
)

// #include <stdlib.h>
// #include <stddef.h>
// #include <stdint.h>
import "C"

type strBuffer struct {
	io []byte
}

var dataStrOutput = &strBuffer{}
var dataStrInput = &strBuffer{}
var connections = make(map[string]net.Conn)

var callBackBufferPtr unsafe.Pointer
var callBackBufferCap int
var workingVM jni.VM
var wg sync.WaitGroup

var profile Profile
var profiles []ShowProfile

func main() {
	println("main started")
}

//export Java_com_dsnteam_dsn_CoreManager_register
func Java_com_dsnteam_dsn_CoreManager_register(env uintptr, _ uintptr, usernameIn uintptr, passwordIn uintptr) {
	username := string(jni.Env(env).GetStringUTF(usernameIn))
	password := string(jni.Env(env).GetStringUTF(passwordIn))
	key := genProfileKey()
	profile = Profile{username: username, password: password, privateKey: key}
	addProfile(profile)
}

//export Java_com_dsnteam_dsn_CoreManager_loadProfiles
func Java_com_dsnteam_dsn_CoreManager_loadProfiles(env uintptr, _ uintptr) {
	profiles = getProfiles()
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesIds
func Java_com_dsnteam_dsn_CoreManager_getProfilesIds(env uintptr, _ uintptr) (ids uintptr) {
	var profilesIds []int
	ids = jni.Env(env).NewIntArray(len(profilesIds))
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetIntArrayElement(ids, i, profiles[i].id)
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesNames
func Java_com_dsnteam_dsn_CoreManager_getProfilesNames(env uintptr, _ uintptr) (usernames uintptr) {
	var profilesStr []string
	dataType := jni.Env(env).FindClass("[Ljava/lang/String;")
	usernames = jni.Env(env).NewObjectArray(len(profilesStr), dataType, 0)
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetObjectArrayElement(usernames, i, jni.Env(env).NewString(profiles[i].username))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_login
func Java_com_dsnteam_dsn_CoreManager_login(env uintptr, _ uintptr, pos int, passwordIn uintptr) {
	password := string(jni.Env(env).GetStringUTF(passwordIn))
	var privateKeyEncBytes []byte
	profile.username, profile.address, privateKeyEncBytes = getProfileByID(profiles[pos].id)
	decProfileKey(privateKeyEncBytes, password)
}

//export Java_com_dsnteam_dsn_CoreManager_runClient
func Java_com_dsnteam_dsn_CoreManager_runClient(env uintptr, _ uintptr, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))
	println("env run client:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}
	con, _ := net.Dial("tcp", address)
	wg.Add(2)
	if _, ok := connections[address]; !ok {
		connections[con.RemoteAddr().String()] = con
	}
	wg.Done()
	go handleConnection(con)
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
		go handleConnection(con)
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

//export Java_com_dsnteam_dsn_CoreManager_setCallBackBuffer
func Java_com_dsnteam_dsn_CoreManager_setCallBackBuffer(env uintptr, _ uintptr, jniBuffer uintptr) {
	callBackBufferPtr = jni.Env(env).GetDirectBufferAddress(jniBuffer)
	callBackBufferCap = jni.Env(env).GetDirectBufferCapacity(jniBuffer)
}

//Symmetrical connection for TCP between f2f
func handleConnection(con net.Conn) {
	defer func(con net.Conn) {
		err := con.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(con)
	println("handling")

	clientReader := bufio.NewReader(con)
	//clientWriter := bufio.NewWriter(con)
	wg.Add(1)
	connections[con.RemoteAddr().String()] = con
	wg.Done()
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
		dataStrOutput.io, err = clientReader.Peek(int(count))
		ErrHandler(err)
		_, err = clientReader.Discard(int(count))
		switch err {
		case nil:
			log.Println(dataStrOutput.io)
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}
		// Responding to the client request
		updateCall(int(count))
		/*if _, err = clientWriter.Write([]byte("Accepted\n")); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}*/
	}
}

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, _ uintptr, inBuffer uintptr, lenIn int, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))

	println(address)
	wg.Add(len(connections))
	for k := range connections {
		println(k)
	}
	wg.Done()

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
		wg.Add(1)
		if _, err = connections[address].Write(bytes); err != nil {
			log.Printf("failed to send the client request: %v\n", err)
		}
		wg.Done()
	case io.EOF:
		log.Println("client closed the connection")
		return
	default:
		log.Printf("client error: %v\n", err)
		return
	}

	// Waiting for the server response
	//var serverResponse []byte
	//serverReader := bufio.NewReader(connections[address])
	//serverResponse, err = serverReader.ReadBytes('\n')

	//switch err {
	//case nil:
	//	log.Println(serverResponse)
	//case io.EOF:
	//	log.Println("server closed the connection")
	//	return
	//default:
	//	log.Printf("server error: %v\n", err)
	//	return
	//}
}

//export Java_com_dsnteam_dsn_CoreManager_exportBytes
func Java_com_dsnteam_dsn_CoreManager_exportBytes(env uintptr, clazz uintptr) uintptr {
	println("env export:", env)
	buffer := jni.Env(env).NewDirectByteBuffer(unsafe.Pointer(&dataStrOutput.io), len(dataStrOutput.io))
	return buffer
}

//Realisation for platform
func updateCall(count int) {
	//Call Application to read structure and update internal data interpretations, update UI.
	var env jni.Env
	env, _ = workingVM.AttachCurrentThread()
	//Test
	println(dataStrOutput.io)
	println("WorkingEnv:", env)
	classInput := env.FindClass("com/dsnteam/dsn/CoreManager")
	println("class_input:", classInput)
	methodId := env.GetStaticMethodID(classInput, "getUpdateCallBack", "(I)V")
	println("MethodID:", methodId)
	var bData []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&bData))
	sh.Data = uintptr(callBackBufferPtr)
	sh.Cap = callBackBufferCap
	sh.Len = len(dataStrOutput.io)
	println("buffer pointer:", callBackBufferPtr)
	copy(bData, dataStrOutput.io)
	println("buffer write done")
	env.CallStaticVoidMethodA(classInput, methodId, jni.Jvalue(count))
	workingVM.DetachCurrentThread()
	runtime.KeepAlive(bData)
}