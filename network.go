package core

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/DSN-team/core/utils"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"runtime"
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

		profilePublicKey := DecodeKey(cur.GetProfilePublicKey())

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

	targetId := user.ID
	if _, ok := cur.Connections.Load(targetId); !ok {
		log.Println("connection not found adding...")
		con, err := net.Dial("tcp", user.Address)
		publicKey := DecodeKey(cur.GetProfilePublicKey())
		_, err = con.Write(publicKey)
		ErrHandler(err)
		cur.Connections.Store(targetId, con)
		go cur.handleRequest(int(targetId), con)
		log.Println("connected to target", targetId)
	} else {
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

func (cur *Profile) WriteRequest(user User, request []byte) {
	var con net.Conn
	if _, ok := cur.Connections.Load(user.ID); !ok {
		log.Println("Not connected to:", user.Username)
		cur.connect(user)
	}
	value, _ := cur.Connections.Load(user.ID)
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
		case utils.RequestData:
			{
				cur.dataHandler(clientId, clientReader)
				break
			}
		case utils.RequestNetwork:
			{
				cur.networkHandler(clientReader)
				break
			}
		case utils.RequestDataVerification:
			{
				cur.verificationHandler(clientId, clientReader)
				break
			}
		case utils.RequestError:
			{
				fmt.Println("Request error")
				os.Exit(-1)
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
	cur.DataStrOutput = append([]byte{utils.RequestData}, cur.DataStrOutput...)
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

	var request FriendRequest
	var requestEncryptMeta FriendRequestMeta
	var requestEncryptSign FriendRequestSign

	bufferSize := utils.GetUint16Reader(clientReader)
	//println("", bufferSize)
	buffer, err := utils.GetBytes(clientReader, uint64(bufferSize))
	ErrHandler(err)
	bufferStream := bytes.NewBuffer(buffer)
	requestDecoder := gob.NewDecoder(bufferStream)
	requestDecoder.Decode(&request)
	signData := cur.decryptAES(request.SignEncrypted)
	signDataStream := bytes.NewBuffer(signData)
	signDecoder := gob.NewDecoder(signDataStream)
	signDecoder.Decode(&requestEncryptSign)

	r := big.NewInt(requestEncryptSign.SignR)
	s := big.NewInt(requestEncryptSign.SignS)
	if cur.verifyData(request.MetaDataEncrypted, *r, *s) == true {
		metaData := cur.decryptAES(request.MetaDataEncrypted)
		metaDataStream := bytes.NewBuffer(metaData)
		metaDataDecoder := gob.NewDecoder(metaDataStream)
		metaDataDecoder.Decode(&requestEncryptMeta)

		if cur.Username == requestEncryptMeta.Username {
			key := UnmarshalPublicKey(requestEncryptSign.FromPublicKey)
			friend = User{Username: requestEncryptMeta.Username, PublicKey: &key, IsFriend: false}
			cur.addUser(&friend)
			cur.addFriendRequest(friend.ID, 1)

			fmt.Println("Friend request done, request from:", requestEncryptMeta.FromUsername, "Accept?")
			cur.DataStrOutput = append([]byte{utils.RequestNetwork}, requestEncryptMeta.FromUsername...)
			cur.DataStrOutput = append(cur.DataStrOutput, requestEncryptSign.FromPublicKey...)
			cur.DataStrOutput = append(cur.DataStrOutput, request.BackTrace...)

			UpdateUI(len(requestEncryptMeta.Username), int(friend.ID))
			return
		}
	}

	request.Depth--
	//Required: Friends.ping && Friends.is_online
	if request.Depth > 0 {
		//encrypted := make([]byte, 0)
		//cur.buildEncryptedPart(&encrypted, request.FromPublicKey, signData, request.MetaDataEncrypted)
		cur.writeFindFriendRequestSecondary(request, int(friend.ID))
	}
}
