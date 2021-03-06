package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/DSN-team/core/utils"
	"io"
	"log"
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
		fmt.Println("Accepted server client")

		profilePublicKey := DecodeKey(cur.GetProfilePublicKey())

		var request Request

		requestDecoder := gob.NewDecoder(con)
		err = requestDecoder.Decode(&request)
		ErrHandler(err)

		clientPublicKeyString := EncodeKey(request.PublicKey)
		profilePublicKeyString := EncodeKey(profilePublicKey)

		user := cur.getUserByPublicKey(clientPublicKeyString)
		if user.ID == 0 {
			user.PublicKeyString = clientPublicKeyString
		}

		if profilePublicKeyString == user.PublicKeyString {
			return
		}

		fmt.Println("connected:", user.ID, user.PublicKeyString)

		if _, ok := cur.Connections.Load(user.PublicKeyString); !ok {
			fmt.Println("connection not found adding...")
			cur.Connections.Store(user.PublicKeyString, con)
		} else {
			fmt.Println("connection already connected")
			return
		}

		go cur.handleRequest(user, con)
	}
}

func (cur *Profile) connect(user User) {
	fmt.Println("Connecting to:", user)
	if _, ok := cur.Connections.Load(user.PublicKeyString); !ok {
		fmt.Println("Connection not found adding...")
		con, err := net.Dial("tcp", user.Address)
		if ErrHandler(err) {
			fmt.Println("Exiting at error")
			return
		}
		var requestBuffer bytes.Buffer
		publicKey := DecodeKey(cur.GetProfilePublicKey())
		request := Request{RequestType: utils.RequestHello, PublicKey: publicKey}

		requestEncoder := gob.NewEncoder(&requestBuffer)
		err = requestEncoder.Encode(&request)
		ErrHandler(err)

		_, err = con.Write(requestBuffer.Bytes())

		ErrHandler(err)
		cur.Connections.Store(user.PublicKeyString, con)
		go cur.handleRequest(user, con)
		fmt.Println("connected to target", user.Username)
	} else {
		return
	}
}

func (cur *Profile) RunServer(address string) {
	go cur.server(address)
}

func (cur *Profile) BuildDataMessage(data []byte, userId uint) (output []byte) {
	fmt.Println("Building data request, data:", data, "user id:", userId)
	dataMessage := DataMessage{string(data)}
	var dataMessageBuffer bytes.Buffer
	dataMessageEncoder := gob.NewEncoder(&dataMessageBuffer)
	err := dataMessageEncoder.Encode(&dataMessage)
	if ErrHandler(err) {
		return
	}
	friendPos, _ := cur.FriendsIDXs.Load(userId)
	output = cur.encryptAES(cur.Friends[friendPos.(int)].PublicKey, dataMessageBuffer.Bytes())
	return
}

func (cur *Profile) WriteRequest(user User, request Request) {
	var con net.Conn
	if _, ok := cur.Connections.Load(user.PublicKeyString); !ok {
		log.Println("Not connected to:", user.Username)
		cur.connect(user)
	}

	value, _ := cur.Connections.Load(user.PublicKeyString)
	if value == nil {
		return
	}
	con = value.(net.Conn)
	runtime.KeepAlive(cur.DataStrInput)
	fmt.Println("writing to:", con.RemoteAddr())

	var requestBuffer bytes.Buffer
	requestEncoder := gob.NewEncoder(&requestBuffer)
	err := requestEncoder.Encode(&request)
	if ErrHandler(err) {
		return
	}

	switch err {
	case nil:
		if _, err = con.Write(requestBuffer.Bytes()); err != nil {
			log.Printf("failed to send the client request: %v\n", err)
		}
		fmt.Println("request sent")
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
	fmt.Println("Handling")
	defer func(con net.Conn) {
		err := con.Close()
		ErrHandler(err)
	}(con)
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
				cur.networkHandler(user, request.Data, false)
				break
			}
		case utils.RequestDataVerification:
			{
				cur.verificationHandler(user, request.Data)
				break
			}
		case utils.RequestAnswer:
			{
				cur.networkHandler(user, request.Data, true)
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
	fmt.Println("Handling data request")
	if user.ID == 0 {
		return
	}

	decryptedData := cur.decryptAES(user.PublicKey, data)

	var dataMessage DataMessage
	dataMessageDecoder := gob.NewDecoder(bytes.NewReader(decryptedData))
	err := dataMessageDecoder.Decode(&dataMessage)
	if ErrHandler(err) {
		return
	}
	cur.DataStrOutput = []byte(dataMessage.Text)
	cur.DataStrOutput = append([]byte{utils.RequestData}, cur.DataStrOutput...)

	fmt.Println("Updating callback")
	UpdateUI(len(dataMessage.Text), int(user.ID))
}

func (cur *Profile) verificationHandler(user User, data []byte) {
	if user.ID <= 0 {
		return
	}

	var dataPing DataPing
	dataPingDecoder := gob.NewDecoder(bytes.NewReader(data))
	err := dataPingDecoder.Decode(&dataPing)
	if ErrHandler(err) {
		return
	}
	cur.Friends[cur.getFriendNumber(user.ID)].Ping = dataPing.Ping
}

func (cur *Profile) networkHandler(user User, data []byte, answer bool) {
	fmt.Println("Handling network request")
	var friend User

	var request FriendRequest
	var requestEncryptMeta FriendRequestMeta

	bufferStream := bytes.NewBuffer(data)
	requestDecoder := gob.NewDecoder(bufferStream)
	err := requestDecoder.Decode(&request)
	ErrHandler(err)

	publicKey := UnmarshalPublicKey(request.FromPublicKey)

	metaData := cur.decryptAES(&publicKey, request.MetaDataEncrypted)
	metaDataStream := bytes.NewBuffer(metaData)
	metaDataDecoder := gob.NewDecoder(metaDataStream)
	err = metaDataDecoder.Decode(&requestEncryptMeta)
	ErrHandler(err)

	if cur.Username == requestEncryptMeta.ToUsername {
		publicKeyString := EncodeKey(request.FromPublicKey)
		friend = User{ProfileID: cur.ID, Username: requestEncryptMeta.FromUsername, Address: requestEncryptMeta.FromAddress,
			PublicKey:       &publicKey,
			IsFriend:        false,
			PublicKeyString: publicKeyString}
		cur.addUser(&friend)
		if !answer {
			friendRequest := cur.addFriendRequest(friend.ID, 1)
			friendRequest.BackTrace = request.BackTrace
			fmt.Println("Friend request done, request from:", requestEncryptMeta.FromUsername)
			cur.DataStrOutput = append([]byte{utils.RequestNetwork}, requestEncryptMeta.FromUsername...)
			cur.DataStrOutput = append(cur.DataStrOutput, request.FromPublicKey...)
			cur.DataStrOutput = append(cur.DataStrOutput, request.BackTrace...)

			UpdateUI(len(requestEncryptMeta.ToUsername), int(friend.ID))
			return
		} else {
			if cur.searchFriendRequest(friend.ID) {
				fmt.Println("SAVING FRIEND FROM answer")
				friend.IsFriend = true
				db.Save(friend)

				cur.DataStrOutput = append([]byte{utils.RequestAnswer}, requestEncryptMeta.FromUsername...)
				cur.DataStrOutput = append(cur.DataStrOutput, request.FromPublicKey...)
				cur.DataStrOutput = append(cur.DataStrOutput, request.BackTrace...)
				cur.LoadFriends()
				UpdateUI(len(requestEncryptMeta.ToUsername), int(friend.ID))
			}
		}

	}

	request.Depth--
	//Required: Friends.ping && Friends.is_online
	if request.Depth > 0 {
		request.BackTrace = append(request.BackTrace, byte(user.ID))
		cur.writeFindFriendRequestSecondary(request, int(friend.ID))
	}

	if answer && len(request.BackTrace) > 0 {
		last := request.BackTrace[len(request.BackTrace)-1]
		cur.answerFindFriendRequestDirect(request, cur.Friends[last])
	}
}
