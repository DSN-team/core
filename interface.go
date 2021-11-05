package core

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/DSN-team/core/utils"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

var UpdateUI = func(int, int) {}

var Profiles []ShowProfile

func (cur *Profile) Register(username, password, address string) bool {
	key := genProfileKey()
	if key == nil {
		return false
	}
	cur.Username, cur.Password, cur.Address, cur.PrivateKey = username, password, address, key
	log.Println(cur)
	cur.Id = addProfile(cur)
	return true
}

func (cur *Profile) Login(password, address string, pos int) (result bool) {
	var privateKeyEncBytes []byte
	cur.Id = Profiles[pos].Id
	cur.Username, cur.Address, privateKeyEncBytes = getProfileByID(Profiles[pos].Id)
	if address != "" {
		cur.Address = address
		//TODO Here need to update DB
	}
	if privateKeyEncBytes == nil {
		return false
	}
	result = cur.decProfileKey(privateKeyEncBytes, password)
	fmt.Println("Login status:", result)
	return
}
func UsernamePos(username string) int {
	profiles := getProfiles()
	pos := -1
	for i := 0; i < len(profiles); i++ {
		if profiles[i].Username == username {
			pos = i
			break
		}
	}
	return pos
}
func LoadProfiles() int {
	Profiles = getProfiles()
	return len(Profiles)
}

func (cur *Profile) GetProfilePublicKey() string {
	return EncPublicKey(MarshalPublicKey(&cur.PrivateKey.PublicKey))
}

func (cur *Profile) AddFriend(username, address, publicKey string) {
	decryptedPublicKey := UnmarshalPublicKey(DecPublicKey(publicKey))
	id := cur.searchUser(username)
	user := User{Username: username, Address: address, PublicKey: &decryptedPublicKey, IsFriend: true}
	if id == -1 {
		cur.addUser(user)
	} else {
		cur.editUser(id, user)
	}
}

func (cur *Profile) FindFriendRequest(username string) (address, publicKey string) {
	request := make([]byte, 8)
	binary.BigEndian.PutUint64(request, uint64(len(username)))
	return "", ""
}

func (cur *Profile) LoadFriends() int {
	println("Loading Friends from db")
	cur.Friends = cur.getFriends()
	return len(cur.Friends)
}

func (cur *Profile) ConnectToFriends() {
	for i := 0; i < len(cur.Friends); i++ {
		go cur.connect(i)
	}
}

func (cur *Profile) ConnectToFriend(userId int) {
	for i := 0; i < len(cur.Friends); i++ {
		go func(index int) {
			if cur.Friends[index].Id == userId {
				cur.connect(index)
				return
			}
		}(i)
	}
}

func (cur *Profile) RunServer(address string) {
	go cur.server(address)
}
func BuildRequest(requestType byte, size uint64, data []byte) (output []byte) {
	request := make([]byte, 0)
	utils.SetByte(&request, requestType)
	utils.SetUint64(&request, size)
	utils.SetBytes(&request, data)
	return request
}
func (cur *Profile) WriteRequest(userId int, request []byte) {
	var con net.Conn
	if _, ok := cur.Connections.Load(userId); !ok {
		log.Println("Not connected to:", userId)
		return
	}
	value, _ := cur.Connections.Load(userId)
	con = value.(net.Conn)
	runtime.KeepAlive(cur.DataStrInput.Io)
	log.Println("writing to:", con.RemoteAddr())

	log.Println("input:", cur.DataStrInput.Io)
	println("input str:", string(cur.DataStrInput.Io))

	switch err {
	case nil:
		println("ClientSend:", request, " count:", len(request))
		if _, err = con.Write(request); err != nil {
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

//TODO Deprecated
func (cur *Profile) WriteDataRequest(userId, lenIn int) {
	var con net.Conn
	if _, ok := cur.Connections.Load(userId); !ok {
		log.Println("Not connected to:", userId)
		return
	}
	value, _ := cur.Connections.Load(userId)
	con = value.(net.Conn)
	runtime.KeepAlive(cur.DataStrInput.Io)
	log.Println("writing to:", con.RemoteAddr())

	log.Println("input:", cur.DataStrInput.Io)
	println("input str:", string(cur.DataStrInput.Io))

	switch err {
	case nil:
		bytes := BuildRequest(RequestData, uint64(lenIn), cur.DataStrInput.Io[0:lenIn])
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

func (cur *Profile) connect(pos int) {
	log.Println("Connecting to friend:", cur.Friends[pos].Username)
	con, err := net.Dial("tcp", cur.Friends[pos].Address)
	for err != nil {
		con, err = net.Dial("tcp", cur.Friends[pos].Address)
		ErrHandler(err)
		time.Sleep(1 * time.Second)
	}

	publicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)
	_, err = con.Write(publicKey)
	ErrHandler(err)

	targetId := cur.Friends[pos].Id
	if _, ok := cur.Connections.Load(targetId); !ok {
		log.Println("connection not found adding...")
		cur.Connections.Store(targetId, con)
	} else {
		log.Println("connection already connected")
		return
	}

	println("connected to target", targetId)
	go cur.handleRequest(targetId, con)
}

func (cur *Profile) server(address string) {
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

		profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)

		clientReader := bufio.NewReader(con)
		publicKeyLen := len(profilePublicKey)
		println(publicKeyLen)
		clientKey, err := clientReader.Peek(publicKeyLen)
		ErrHandler(err)
		_, err = clientReader.Discard(publicKeyLen)
		ErrHandler(err)

		log.Println("reader size:", clientReader.Size())

		var clientId int

		clientPublicKeyString := EncPublicKey(clientKey)
		profilePublicKeyString := EncPublicKey(profilePublicKey)
		log.Println("Current profile public key:", profilePublicKeyString)
		log.Println("client public key:", clientPublicKeyString)

		if profilePublicKeyString != clientPublicKeyString {
			clientId = getUserByPublicKey(clientPublicKeyString)
			if clientId == 0 {
				log.Println("not found in database")
				return
			}
		}

		log.Println("connected:", clientId, clientPublicKeyString)

		if _, ok := cur.Connections.Load(clientId); !ok {
			log.Println("connection not found adding...")
			cur.Connections.Store(clientId, con)
		} else {
			log.Println("connection already connected")
			return
		}

		go cur.handleRequest(clientId, con)
	}
}

const (
	RequestData    = byte(0)
	RequestNetwork = byte(1)
)

func (cur *Profile) dataHandler(clientId int, clientReader *bufio.Reader) {
	// Waiting for the client request
	count := utils.GetUint64Reader(clientReader)
	log.Println("Count:", count)
	cur.DataStrOutput.Io, err = utils.GetBytes(clientReader, count)
	switch err {
	case nil:
		log.Println(cur.DataStrOutput.Io)
	case io.EOF:
		log.Println("client closed the connection by terminating the process")
		return
	default:
		log.Printf("error: %v\n", err)
		return
	}
	log.Println("updating callback")
	UpdateUI(int(count), clientId)
}
func (cur *Profile) networkHandler(_ int, clientReader *bufio.Reader) {
	//Request size
	count := utils.GetUint64Reader(clientReader)
	userNameSize := utils.GetUint16Reader(clientReader)
	username, _ := utils.GetBytes(clientReader, uint64(userNameSize))
	fmt.Println("count:", count, " userNameSize:", userNameSize, " username:", username)
}

//Symmetrical connection for TCP between f2f
func (cur *Profile) handleRequest(clientId int, con net.Conn) {
	log.Println("handling")
	defer func(con net.Conn) {
		err := con.Close()
		ErrHandler(err)
	}(con)
	clientReader := bufio.NewReader(con)
	for {
		requestType := utils.GetByte(clientReader)
		fmt.Println("Request type:", requestType)
		switch requestType {
		case RequestData:
			{
				cur.dataHandler(clientId, clientReader)
				break
			}
		case RequestNetwork:
			{
				cur.networkHandler(clientId, clientReader)
				break
			}
		}
	}
}
