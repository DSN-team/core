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

var jniUsed = true

func testAES() {
	profileKey := genProfileKey()
	profile.privateKey = profileKey
	publicKey := profileKey.PublicKey
	println("publicKey:", (publicKey).X.Uint64())
	println(string(decryptAES(encryptAES(&publicKey, []byte("12312")))))
}

func main() {
	jniUsed = false
	println("main started")
	testAES()
}
func initDB() {
	startDB()
}
func register(username, password string) bool {
	key := genProfileKey()
	if key == nil {
		return false
	}
	profile = Profile{username: username, password: password, privateKey: key}
	log.Println(profile)
	addProfile(profile)
	return true
}

func login(password string, pos int) (result bool) {
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
func loadProfiles() int {
	profiles = getProfiles()
	return len(profiles)
}

func getProfilePublicKey() string {
	return encPublicKey(marshalPublicKey(&profile.privateKey.PublicKey))
}

func addFriend(address, publicKey string) {
	decryptedPublicKey := unmarshalPublicKey(decPublicKey(publicKey))
	user := User{address: address, publicKey: &decryptedPublicKey, isFriend: true}
	addUser(user)
}

func loadFriends() int {
	println("Loading friends from db")
	friends = getFriends()
	return len(friends)
}

func connectToFriends() {
	for i := 0; i < len(friends); i++ {
		go connect(i)
	}
}

func connectToFriend(userId int) {
	for i := 0; i < len(friends); i++ {
		go func(index int) {
			if friends[index].id == userId {
				connect(index)
				return
			}
		}(i)
	}
}

func runServer(address string) {
	go server(address)
}

//dataStrInput.io must be non null
func writeBytes(userId, lenIn int) {

	var con net.Conn
	if _, ok := connections.Load(userId); !ok {
		log.Println("Not connected to:", userId)
		return
	}
	value, _ := connections.Load(userId)
	con = value.(net.Conn)
	runtime.KeepAlive(dataStrInput.io)
	log.Println("writing to:", con.RemoteAddr())

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

//Realisation for platform
func updateCall(count int, userId int) {
	//Call Application to read structure and update internal data interpretations, update UI.
	var env jni.Env
	env, _ = workingVM.AttachCurrentThread()
	//Test
	println(dataStrOutput.io)
	if jniUsed {
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
}
