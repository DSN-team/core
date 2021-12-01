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

type Request struct {
	RequestType byte
	PublicKey   []byte
	Data        []byte
}

type DataMessage struct {
	Text string
}

type DataPing struct {
	Ping int
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

		user := cur.getUserByPublicKey(clientPublicKeyString)

		if profilePublicKeyString == clientPublicKeyString {
			return
		}

		log.Println("connected:", user.ID, clientPublicKeyString)

		if _, ok := cur.Connections.Load(user.ID); !ok {
			log.Println("connection not found adding...")
			cur.Connections.Store(user.ID, con)
		} else {
			log.Println("connection already connected")
			return
		}

		go cur.handleRequest(user, con)
	}
}

func (cur *Profile) connect(user User) {
	if _, ok := cur.Connections.Load(user.ID); !ok {
		log.Println("connection not found adding...")
		con, err := net.Dial("tcp", user.Address)
		publicKey := DecodeKey(cur.GetProfilePublicKey())
		_, err = con.Write(publicKey)
		ErrHandler(err)
		cur.Connections.Store(user.ID, con)
		go cur.handleRequest(user, con)
		log.Println("connected to target", user.Username)
	} else {
		return
	}
}

func (cur *Profile) RunServer(address string) {
	go cur.server(address)
}

func (cur *Profile) BuildDataMessage(data []byte, userId uint) (output []byte) {
	log.Println("Building data request, data:", data, "user id:", userId)
	dataMessage := DataMessage{string(data)}
	var dataMessageBuffer bytes.Buffer
	dataMessageEncoder := gob.NewEncoder(&dataMessageBuffer)
	dataMessageEncoder.Encode(&dataMessage)
	friendPos, _ := cur.FriendsIDXs.Load(userId)
	output = cur.encryptAES(cur.Friends[friendPos.(int)].PublicKey, dataMessageBuffer.Bytes())
	return
}

func (cur *Profile) WriteRequest(user User, request Request) {
	var con net.Conn
	if _, ok := cur.Connections.Load(user.ID); !ok {
		log.Println("Not connected to:", user.Username)
		cur.connect(user)
	}
	value, _ := cur.Connections.Load(user.ID)
	con = value.(net.Conn)
	runtime.KeepAlive(cur.DataStrInput)
	log.Println("writing to:", con.RemoteAddr())

	var requestBuffer bytes.Buffer
	requestEncoder := gob.NewEncoder(&requestBuffer)
	requestEncoder.Encode(&request)

	switch err {
	case nil:
		if _, err = con.Write(requestBuffer.Bytes()); err != nil {
			log.Printf("failed to send the client request: %v\n", err)
		}
		log.Println("request sent")
	case io.EOF:
		log.Println("client closed the connection")
		return
	default:
		log.Printf("client error: %v\n", err)
		return
	}
}

//Symmetrical connection for TCP between f2f
func (cur *Profile) handleRequest(user User, con net.Conn) {
	log.Println("handling")
	defer func(con net.Conn) {
		err := con.Close()
		ErrHandler(err)
	}(con)
	//clientReader := bufio.NewReader(con)
	for {
		requestDecoder := gob.NewDecoder(con)
		var request Request
		err := requestDecoder.Decode(&request)
		ErrHandler(err)
		fmt.Println("Request type:", request.RequestType)
		switch request.RequestType {
		case utils.RequestData:
			{
				cur.dataHandler(user, request.Data)
				break
			}
		case utils.RequestNetwork:
			{
				cur.networkHandler(request.Data)
				break
			}
		case utils.RequestDataVerification:
			{
				cur.verificationHandler(user, request.Data)
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

func (cur *Profile) dataHandler(user User, data []byte) {
	log.Println("handling data request")
	if user.ID == 0 {
		return
	}

	decryptedData := cur.decryptAES(user.PublicKey, data)

	var dataMessage DataMessage
	dataMessageDecoder := gob.NewDecoder(bytes.NewReader(decryptedData))
	dataMessageDecoder.Decode(&dataMessage)
	cur.DataStrOutput = []byte(dataMessage.Text)
	cur.DataStrOutput = append([]byte{utils.RequestData}, cur.DataStrOutput...)

	log.Println("updating callback")
	UpdateUI(len(dataMessage.Text), int(user.ID))
}

func (cur *Profile) verificationHandler(user User, data []byte) {
	if user.ID <= 0 {
		return
	}

	var dataPing DataPing
	dataPingDecoder := gob.NewDecoder(bytes.NewReader(data))
	dataPingDecoder.Decode(&dataPing)
	cur.Friends[cur.getFriendNumber(int(user.ID))].Ping = dataPing.Ping
}

func (cur *Profile) networkHandler(data []byte) {
	var friend User

	var request FriendRequest
	var requestEncryptMeta FriendRequestMeta
	var requestEncryptSign FriendRequestSign

	bufferStream := bytes.NewBuffer(data)
	requestDecoder := gob.NewDecoder(bufferStream)
	requestDecoder.Decode(&request)

	publicKey := UnmarshalPublicKey(request.FromPublicKey)

	signData := cur.decryptAES(&publicKey, request.SignEncrypted)
	signDataStream := bytes.NewBuffer(signData)
	signDecoder := gob.NewDecoder(signDataStream)
	signDecoder.Decode(&requestEncryptSign)

	r := big.NewInt(requestEncryptSign.SignR)
	s := big.NewInt(requestEncryptSign.SignS)
	if cur.verifyData(request.MetaDataEncrypted, *r, *s) == true {
		metaData := cur.decryptAES(&publicKey, request.MetaDataEncrypted)
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
		cur.writeFindFriendRequestSecondary(request, int(friend.ID))
	}
}
