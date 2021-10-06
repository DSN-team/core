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
var connections = &sync.Map{}

var callBackBufferPtr unsafe.Pointer
var callBackBufferCap int
var workingVM jni.VM

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
	log.Println(profile)
	addProfile(profile)
	return true
}

//export Java_com_dsnteam_dsn_CoreManager_login
func Java_com_dsnteam_dsn_CoreManager_login(env uintptr, _ uintptr, pos int, passwordIn uintptr) (result bool) {
	password := string(jni.Env(env).GetStringUTF(passwordIn))
	var privateKeyEncBytes []byte
	profile.id = profiles[pos].id
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
	return jni.Env(env).NewString(encPublicKey(marshalPublicKey(&profile.privateKey.PublicKey)))
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
	decryptedPublicKey := unmarshalPublicKey(decPublicKey(publicKey))
	user := User{address: address, publicKey: &decryptedPublicKey, isFriend: true}
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
	for i := 0; i < len(friends); i++ {
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
		jni.Env(env).SetObjectArrayElement(publicKey, i, jni.Env(env).NewString(encPublicKey(marshalPublicKey(friends[i].publicKey))))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_connectToFriends
func Java_com_dsnteam_dsn_CoreManager_connectToFriends(env uintptr, _ uintptr) {
	for i := 0; i < len(friends); i++ {
		go connect(i)
	}
}

//export Java_com_dsnteam_dsn_CoreManager_connectToFriend
func Java_com_dsnteam_dsn_CoreManager_connectToFriend(env uintptr, _ uintptr, userId int) {
	for i := 0; i < len(friends); i++ {
		go func(index int) {
			if friends[index].id == userId {
				connect(index)
				return
			}
		}(i)
	}
}

func connect(pos int) {

	con, err := net.Dial("tcp", friends[pos].address)
	for err != nil {
		con, err = net.Dial("tcp", friends[pos].address)
		ErrHandler(err)
		time.Sleep(1 * time.Second)
	}

	publicKey := marshalPublicKey(&profile.privateKey.PublicKey)
	_, err = con.Write(publicKey)
	ErrHandler(err)

	targetId := friends[pos].id

	if _, ok := connections.Load(targetId); !ok {
		log.Println("connection not found adding...")
		connections.Store(targetId, con)
	} else {
		log.Println("connection already connected")
		return
	}

	println("connected to target", targetId)
	go handleConnection(targetId, con)
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
		println("accepted server client")

		profilePublicKey := marshalPublicKey(&profile.privateKey.PublicKey)

		clientReader := bufio.NewReader(con)
		publicKeyLen := len(profilePublicKey)
		println(publicKeyLen)
		clientKey, err := clientReader.Peek(publicKeyLen)
		ErrHandler(err)
		_, err = clientReader.Discard(publicKeyLen)
		ErrHandler(err)

		log.Println("reader size:", clientReader.Size())

		var clientId int

		clientPublicKeyString := encPublicKey(clientKey)
		profilePublicKeyString := encPublicKey(profilePublicKey)
		log.Println("profile public key:", profilePublicKeyString)
		log.Println("client public key:", clientPublicKeyString)

		if profilePublicKeyString != clientPublicKeyString {
			clientId = getUserByPublicKey(clientPublicKeyString)
			if clientId == 0 {
				log.Println("not found in database")
				return
			}
		}

		log.Println("connected:", clientId, clientPublicKeyString)

		if _, ok := connections.Load(clientId); !ok {
			log.Println("connection not found adding...")
			connections.Store(clientId, con)
		} else {
			log.Println("connection already connected")
			return
		}

		go handleConnection(clientId, con)
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
func handleConnection(clientId int, con net.Conn) {
	log.Println("handling")

	defer func(con net.Conn) {
		err := con.Close()
		ErrHandler(err)
	}(con)

	clientReader := bufio.NewReader(con)

	log.Println("bufio")
	for {
		// Waiting for the client request
		log.Println("reading")
		state, err := clientReader.Peek(9)
		ErrHandler(err)
		_, err = clientReader.Discard(9)
		ErrHandler(err)
		count := binary.BigEndian.Uint64(state[0:8])
		log.Println("Count:", count)
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

		log.Println("updating callback")
		updateCall(int(count), clientId)
	}
}

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, _ uintptr, inBuffer uintptr, lenIn int, userId int) {
	var con net.Conn
	if _, ok := connections.Load(userId); !ok {
		log.Println("Not connected to:", userId)
		return
	}
	value, _ := connections.Load(userId)
	con = value.(net.Conn)

	log.Println("writing to:", con.RemoteAddr())

	log.Println("env write:", env)
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

		if _, err = con.Write(bytes); err != nil {
			log.Printf("failed to send the client request: %v\n", err)
		}

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
	methodId := env.GetStaticMethodID(classInput, "getUpdateCallBack", "(II)V")
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
