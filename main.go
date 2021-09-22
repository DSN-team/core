package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/ClarkGuan/jni"
	"io"
	"log"
	"net"
	"reflect"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// #include <stdlib.h>
// #include <stddef.h>
// #include <stdint.h>
import "C"

//Include for correct jni

type strBuffer struct {
	io []byte
}

var dataStrOutput = &strBuffer{}
var dataStrInput = &strBuffer{}
var connections = make(map[int]net.Conn)

var callBackBufferPtr unsafe.Pointer
var callBackBufferCap int
var workingVM jni.VM
var wg sync.WaitGroup

var profile Profile
var profiles []ShowProfile
var friends []User

func main() {
	println("main started")
}

//export Java_com_dsnteam_dsn_CoreManager_initDB
func Java_com_dsnteam_dsn_CoreManager_initDB(env uintptr, _ uintptr) {
	startDB()
}

//export Java_com_dsnteam_dsn_CoreManager_register
func Java_com_dsnteam_dsn_CoreManager_register(env uintptr, _ uintptr, usernameIn uintptr, passwordIn uintptr) (result bool) {
	username, password := string(jni.Env(env).GetStringUTF(usernameIn)), string(jni.Env(env).GetStringUTF(passwordIn))
	key := genProfileKey()
	if key == nil {
		return false
	}
	profile = Profile{username: username, password: password, privateKey: key}
	addProfile(profile)
	return true
}

//export Java_com_dsnteam_dsn_CoreManager_login
func Java_com_dsnteam_dsn_CoreManager_login(env uintptr, _ uintptr, pos int, passwordIn uintptr) (result bool) {
	password := string(jni.Env(env).GetStringUTF(passwordIn))
	var privateKeyEncBytes []byte
	profile.username, profile.address, privateKeyEncBytes = getProfileByID(profiles[pos].id)
	if privateKeyEncBytes == nil {
		return false
	}
	result = decProfileKey(privateKeyEncBytes, password)
	fmt.Println("login status:", result)
	return
}

//export Java_com_dsnteam_dsn_CoreManager_loadProfiles
func Java_com_dsnteam_dsn_CoreManager_loadProfiles(env uintptr, _ uintptr) int {
	profiles = getProfiles()
	return len(profiles)
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesIds
func Java_com_dsnteam_dsn_CoreManager_getProfilesIds(env uintptr, _ uintptr) (ids uintptr) {
	ids = jni.Env(env).NewIntArray(len(profiles))
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetIntArrayElement(ids, i, profiles[i].id)
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesNames
func Java_com_dsnteam_dsn_CoreManager_getProfilesNames(env uintptr, _ uintptr) (usernames uintptr) {
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	usernames = jni.Env(env).NewObjectArray(len(profiles), dataType, 0)
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetObjectArrayElement(usernames, i, jni.Env(env).NewString(profiles[i].username))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilePublicKey
func Java_com_dsnteam_dsn_CoreManager_getProfilePublicKey(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(encPublicKey(profile.privateKey.PublicKey))
}

//export Java_com_dsnteam_dsn_CoreManager_getProfileName
func Java_com_dsnteam_dsn_CoreManager_getProfileName(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(profile.username)
}

//export Java_com_dsnteam_dsn_CoreManager_getProfileAddress
func Java_com_dsnteam_dsn_CoreManager_getProfileAddress(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(profile.address)
}

//export Java_com_dsnteam_dsn_CoreManager_addFriend
func Java_com_dsnteam_dsn_CoreManager_addFriend(env uintptr, _ uintptr, addressIn uintptr, publicKeyIn uintptr) {
	address, publicKey := string(jni.Env(env).GetStringUTF(addressIn)), string(jni.Env(env).GetStringUTF(publicKeyIn))
	user := User{address: address, publicKey: decPublicKey(publicKey), isFriend: true}
	addUser(user)
}

//export Java_com_dsnteam_dsn_CoreManager_loadFriends
func Java_com_dsnteam_dsn_CoreManager_loadFriends(env uintptr, _ uintptr) int {
	println("Loading friends from db")
	friends = getFriends()
	return len(friends)
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsIds
func Java_com_dsnteam_dsn_CoreManager_getFriendsIds(env uintptr, _ uintptr) (ids uintptr) {
	ids = jni.Env(env).NewIntArray(len(friends))
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetIntArrayElement(ids, i, friends[i].id)
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsNames
func Java_com_dsnteam_dsn_CoreManager_getFriendsNames(env uintptr, _ uintptr) (usernames uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	usernames = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(usernames, i, jni.Env(env).NewString(friends[i].username))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsAddresses
func Java_com_dsnteam_dsn_CoreManager_getFriendsAddresses(env uintptr, _ uintptr) (address uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	address = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(address, i, jni.Env(env).NewString(friends[i].address))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsPublicKeys
func Java_com_dsnteam_dsn_CoreManager_getFriendsPublicKeys(env uintptr, _ uintptr) (publicKey uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	publicKey = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(publicKey, i, jni.Env(env).NewString(encPublicKey(*friends[i].publicKey)))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_connectToFriends
func Java_com_dsnteam_dsn_CoreManager_connectToFriends(env uintptr, ptr uintptr) {
	for i := 0; i < len(friends); i++ {
		go Java_com_dsnteam_dsn_CoreManager_runClient(env, ptr, i)
	}
}

//export Java_com_dsnteam_dsn_CoreManager_runClient
func Java_com_dsnteam_dsn_CoreManager_runClient(env uintptr, _ uintptr, pos int) {
	//println("env run client:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}

	con, err := net.Dial("tcp", friends[pos].address)
	for err != nil {
		con, err = net.Dial("tcp", friends[pos].address)
		ErrHandler(err)
		time.Sleep(100 * time.Millisecond)
	}

	publicKey := marshalPublicKey(profile.privateKey.PublicKey)
	_, err = con.Write(publicKey)
	ErrHandler(err)
	println("public key size:", len(publicKey))
	go handleConnection(con)
}

func server(address string) {
	ln, err := net.Listen("tcp", address)
	ErrHandler(err)
	defer func(ln net.Listener) {
		err := ln.Close()
		ErrHandler(err)
	}(ln)
	ErrHandler(err)
	for {
		con, err := ln.Accept()
		ErrHandler(err)
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
		ErrHandler(err)
	}(con)
	println("handling")

	clientReader := bufio.NewReader(con)
	publicKeyLen := len(marshalPublicKey(profile.privateKey.PublicKey))
	key, err := clientReader.Peek(publicKeyLen)
	ErrHandler(err)
	_, err = clientReader.Discard(publicKeyLen)
	ErrHandler(err)

	publicKeyString := encPublicKey(*unmarshalPublicKey(key))

	userId := getUserByPublicKey(publicKeyString)

	wg.Add(2)
	if _, ok := connections[userId]; !ok {
		connections[userId] = con
	}
	wg.Done()

	println(con.RemoteAddr().String())
	println("bufio")
	for {
		// Waiting for the client request
		println("reading")
		state, err := clientReader.Peek(9)
		ErrHandler(err)
		_, err = clientReader.Discard(9)
		ErrHandler(err)
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

		updateCall(int(count), userId)
	}
}

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, _ uintptr, inBuffer uintptr, lenIn int, userId int) {
	println("env write:", env)
	defer runtime.KeepAlive(dataStrInput.io)
	point, size := jni.Env(env).GetDirectBufferAddress(inBuffer), jni.Env(env).GetDirectBufferCapacity(inBuffer)

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dataStrInput.io))
	sh.Data = uintptr(point)
	sh.Len = lenIn
	sh.Cap = size

	runtime.KeepAlive(dataStrInput.io)
	log.Println("input:", dataStrInput.io)
	println("input str:", string(dataStrInput.io))

	switch err {
	case nil:
		bs := make([]byte, 9)
		binary.BigEndian.PutUint64(bs, uint64(lenIn))
		bs[8] = '\n'
		bytes := append(bs, dataStrInput.io...)
		println("ClientSend:", bytes, " count:", lenIn)
		wg.Add(1)
		if _, err = connections[userId].Write(bytes); err != nil {
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
}

//Realisation for platform
func updateCall(count int, userId int) {
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
	env.CallStaticVoidMethodA(classInput, methodId, jni.Jvalue(count), jni.Jvalue(userId))
	workingVM.DetachCurrentThread()
	runtime.KeepAlive(bData)
}
