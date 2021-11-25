package core

import (
	"bufio"
	"fmt"
	"github.com/DSN-team/core/utils"
	"io"
	"log"
	"math/big"
	"net"
	"runtime"
)

const (
	RequestData             = byte(0)
	RequestDataVerification = byte(1)
	RequestNetwork          = byte(2)
)

type NetworkInterface interface {
	startTimer()
	sendData(callback func())
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
		log.Println("accepted server client")

		profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)

		clientReader := bufio.NewReader(con)
		publicKeyLen := len(profilePublicKey)
		//TODO Add network type
		log.Println(publicKeyLen)
		clientKey, err := clientReader.Peek(publicKeyLen)
		ErrHandler(err)
		_, err = clientReader.Discard(publicKeyLen)
		ErrHandler(err)

		log.Println("reader size:", clientReader.Size())

		clientPublicKeyString := EncodeKey(clientKey)
		profilePublicKeyString := EncodeKey(profilePublicKey)

		var clientId int

		if profilePublicKeyString != clientPublicKeyString {
			user := cur.getUserByPublicKey(clientPublicKeyString)
			clientId = int(user.ID)
			if clientId == 0 {
				log.Println("not found in database")
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

func (cur *Profile) connect(user User) {
	log.Println("Connecting to friend:", user.Username)

	targetId := int(user.ID)
	if _, ok := cur.Connections.Load(targetId); !ok {
		log.Println("connection not found adding...")
		con, err := net.Dial("tcp", user.Address)
		publicKey := DecodeKey(cur.GetProfilePublicKey())
		_, err = con.Write(publicKey)
		ErrHandler(err)
		cur.Connections.Store(targetId, con)
		go cur.handleRequest(targetId, con)
		log.Println("connected to target", targetId)
	} else {
		log.Println("connection already connected")
		return
	}
}

func (cur *Profile) RunServer(address string) {
	go cur.server(address)
}

func (cur *Profile) BuildDataRequest(requestType byte, size uint64, data []byte, userId uint) (output []byte) {
	log.Println("Building data request, request type: ", requestType, "size:",
		size, "data:", data, "user id:", userId)
	request := make([]byte, 0)
	utils.SetByte(&request, requestType)
	utils.SetUint64(&request, size)
	friendPos, _ := cur.FriendsIDXs.Load(userId)
	utils.SetBytes(&request, cur.encryptAES(cur.Friends[friendPos.(int)].PublicKey, data))
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
	runtime.KeepAlive(cur.DataStrInput)
	log.Println("writing to:", con.RemoteAddr())

	log.Println("input:", cur.DataStrInput)
	log.Println("input str:", string(cur.DataStrInput))

	switch err {
	case nil:
		log.Println("ClientSend:", request, " count:", len(request))
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
				cur.networkHandler(clientReader)
				break
			}
		case RequestDataVerification:
			{
				cur.verificationHandler(clientId, clientReader)
				break
			}
		}
	}
}

func (cur *Profile) dataHandler(clientId int, clientReader *bufio.Reader) {
	if clientId == -1 {
		return
	}

	// Waiting for the client request
	count := utils.GetUint64Reader(clientReader)
	log.Println("Count:", count)
	encData, err := utils.GetBytes(clientReader, count)
	cur.DataStrOutput = cur.decryptAES(encData)
	cur.DataStrOutput = append([]byte{RequestData}, cur.DataStrOutput...)
	switch err {
	case nil:
		log.Println(cur.DataStrOutput)
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

func (cur *Profile) verificationHandler(clientId int, clientReader *bufio.Reader) {
	if clientId == -1 {
		return
	}
	cur.Friends[cur.getFriendNumber(clientId)].Ping = int(utils.GetUint16Reader(clientReader))
}

func (cur *Profile) networkHandler(clientReader *bufio.Reader) {
	var friend User
	//metaData sizes
	requestDepth := utils.GetUint8Reader(clientReader)
	requestDegree := utils.GetUint8Reader(clientReader)
	backTraceSize := utils.GetUint8Reader(clientReader)
	backTrace, _ := utils.GetBytes(clientReader, uint64(backTraceSize))

	userNameSize := utils.GetUint16Reader(clientReader)
	fromUserNameSize := utils.GetUint16Reader(clientReader)

	//todo complete this
	publicKey, _ := utils.GetBytes(clientReader, uint64(128))
	metaDataSize := utils.GetUint32Reader(clientReader)
	metaDataEncrypted, _ := utils.GetBytes(clientReader, uint64(metaDataSize))
	signSize := utils.GetUint32Reader(clientReader)
	signDataEncrypted, _ := utils.GetBytes(clientReader, uint64(signSize))

	signData := cur.decryptAES(signDataEncrypted)

	r := new(big.Int).SetBytes(signData[0 : signSize/2])
	s := new(big.Int).SetBytes(signData[signSize/2 : signSize/2])

	if cur.verifyData(metaDataEncrypted, *r, *s) == true {
		metaData := cur.decryptAES(metaDataEncrypted)
		username := metaData[0:userNameSize]
		fromUsername := metaData[userNameSize:fromUserNameSize]

		fmt.Println("UserNameSize:", userNameSize, " FromUserNameSize:", fromUserNameSize, " Username:", username,
			" Depth:", requestDepth, " BackTrace:", backTrace)
		if cur.Username == string(username) {
			key := UnmarshalPublicKey(publicKey)
			friend = User{Username: string(username), PublicKey: &key, IsFriend: false}
			cur.addUser(&friend)
			cur.addFriendRequest(friend.ID, 1)

			fmt.Println("Friend request done, request from:", string(fromUsername), "Accept?")
			cur.DataStrOutput = append([]byte{RequestNetwork}, fromUsername...)
			cur.DataStrOutput = append(cur.DataStrOutput, publicKey...)
			cur.DataStrOutput = append(cur.DataStrOutput, backTrace...)

			UpdateUI(int(userNameSize), int(friend.ID))
			return
		}
	}

	requestDepth--
	//Required: Friends.ping && Friends.is_online
	if requestDepth > 0 {
		encrypted := make([]byte, 0)
		cur.buildEncryptedPart(&encrypted, publicKey, signData, metaDataEncrypted)
		cur.writeFindFriendRequestSecondary(int(requestDepth), int(requestDegree), int(friend.ID), backTrace, encrypted)
	}
}
