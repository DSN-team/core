package core

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

type strBuffer struct {
	io []byte
}

var dataStrOutput = &strBuffer{}
var dataStrInput = &strBuffer{}
var connections = &sync.Map{}

var SelectedProfile Profile
var Profiles []ShowProfile
var Friends []User

var jniUsed = true

func testAES() {
	profileKey := genProfileKey()
	SelectedProfile.privateKey = profileKey
	publicKey := profileKey.PublicKey
	println("publicKey:", (publicKey).X.Uint64())
	println(string(decryptAES(encryptAES(&publicKey, []byte("12312")))))
}

func main() {
	jniUsed = false
	println("main started")
	testAES()
}

func Register(username, password string) bool {
	key := genProfileKey()
	if key == nil {
		return false
	}
	SelectedProfile = Profile{username: username, password: password, privateKey: key}
	log.Println(SelectedProfile)
	addProfile(SelectedProfile)
	return true
}

func Login(password string, pos int) (result bool) {
	var privateKeyEncBytes []byte
	SelectedProfile.id = Profiles[pos].id
	SelectedProfile.username, SelectedProfile.address, privateKeyEncBytes = getProfileByID(Profiles[pos].id)
	if privateKeyEncBytes == nil {
		return false
	}
	result = decProfileKey(privateKeyEncBytes, password)
	fmt.Println("Login status:", result)
	return
}
func LoadProfiles() int {
	Profiles = getProfiles()
	return len(Profiles)
}

func GetProfilePublicKey() string {
	return encPublicKey(marshalPublicKey(&SelectedProfile.privateKey.PublicKey))
}

func AddFriend(address, publicKey string) {
	decryptedPublicKey := unmarshalPublicKey(decPublicKey(publicKey))
	user := User{address: address, publicKey: &decryptedPublicKey, isFriend: true}
	addUser(user)
}

func LoadFriends() int {
	println("Loading Friends from db")
	Friends = getFriends()
	return len(Friends)
}

func ConnectToFriends() {
	for i := 0; i < len(Friends); i++ {
		go connect(i)
	}
}

func ConnectToFriend(userId int) {
	for i := 0; i < len(Friends); i++ {
		go func(index int) {
			if Friends[index].id == userId {
				connect(index)
				return
			}
		}(i)
	}
}

func RunServer(address string) {
	go server(address)
}

func WriteBytes(userId, lenIn int) {
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
	con, err := net.Dial("tcp", Friends[pos].address)
	for err != nil {
		con, err = net.Dial("tcp", Friends[pos].address)
		ErrHandler(err)
		time.Sleep(1 * time.Second)
	}

	publicKey := marshalPublicKey(&SelectedProfile.privateKey.PublicKey)
	_, err = con.Write(publicKey)
	ErrHandler(err)

	targetId := Friends[pos].id

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

		profilePublicKey := marshalPublicKey(&SelectedProfile.privateKey.PublicKey)

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
		log.Println("SelectedProfile public key:", profilePublicKeyString)
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
		//updateCall(int(count), clientId)
	}
}
