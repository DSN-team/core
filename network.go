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

func (profile *Profile) server(address string) {
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

		profilePublicKey := MarshalPublicKey(&profile.PrivateKey.PublicKey)

		clientReader := bufio.NewReader(con)
		publicKeyLen := len(profilePublicKey)
		//TODO Add network type
		log.Println(publicKeyLen)
		clientKey, err := clientReader.Peek(publicKeyLen)
		ErrHandler(err)
		_, err = clientReader.Discard(publicKeyLen)
		ErrHandler(err)

		log.Println("reader size:", clientReader.Size())

		clientPublicKeyString := EncPublicKey(clientKey)
		profilePublicKeyString := EncPublicKey(profilePublicKey)

		var clientId uint

		if profilePublicKeyString != clientPublicKeyString {
			clientId = profile.getUserByPublicKey(clientPublicKeyString).ID
			if clientId == -1 {
				log.Println("not found in database")
			}
		}

		log.Println("connected:", clientId, clientPublicKeyString)

		if _, ok := profile.Connections.Load(clientId); !ok {
			log.Println("connection not found adding...")
			profile.Connections.Store(clientId, con)
		} else {
			log.Println("connection already connected")
			return
		}

		go profile.handleRequest(clientId, con)
	}
}

func (profile *Profile) connect(user *User) {
	con, err := net.Dial("tcp", user.Address)
	publicKey := MarshalPublicKey(&profile.PrivateKey.PublicKey)
	_, err = con.Write(publicKey)
	ErrHandler(err)
	targetId := user.ID
	if _, ok := profile.Connections.Load(targetId); !ok {
		log.Println("connection not found adding...")
		profile.Connections.Store(targetId, con)
		go profile.handleRequest(targetId, con)
		log.Println("connected to target", targetId)
	} else {
		return
	}
}

func (profile *Profile) RunServer(address string) {
	go profile.server(address)
}

func (profile *Profile) BuildDataRequest(requestType byte, size uint64, data []byte, userId int) (output []byte) {
	log.Println("Building data Request, Request type: ", requestType, "size:",
		size, "data:", data, "User id:", userId)
	request := make([]byte, 0)
	utils.SetByte(&request, requestType)
	utils.SetUint64(&request, size)
	friendPos, _ := profile.FriendsIDXs.Load(userId)
	utils.SetBytes(&request, profile.encryptAES(profile.Friends[friendPos.(int)].user.PublicKey, data))
	return request
}

func (profile *Profile) WriteRequest(userId uint, request []byte) {
	var con net.Conn
	if _, ok := profile.Connections.Load(userId); !ok {
		log.Println("Not connected to:", userId)
		return
	}
	value, _ := profile.Connections.Load(userId)
	con = value.(net.Conn)
	runtime.KeepAlive(profile.DataStrInput)
	log.Println("writing to:", con.RemoteAddr())

	log.Println("input:", profile.DataStrInput)
	log.Println("input str:", string(profile.DataStrInput))

	switch err {
	case nil:
		log.Println("ClientSend:", request, " count:", len(request))
		if _, err = con.Write(request); err != nil {
			log.Printf("failed to send the client Request: %v\n", err)
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
func (profile *Profile) handleRequest(clientId uint, con net.Conn) {
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
				profile.dataHandler(clientId, clientReader)
				break
			}
		case RequestNetwork:
			{
				profile.networkHandler(clientReader)
				break
			}
		case RequestDataVerification:
			{
				profile.verificationHandler(clientId, clientReader)
				break
			}
		}
	}
}

func (profile *Profile) dataHandler(clientId uint, clientReader *bufio.Reader) {
	if clientId == -1 {
		return
	}

	// Waiting for the client Request
	count := utils.GetUint64Reader(clientReader)
	log.Println("Count:", count)
	encData, err := utils.GetBytes(clientReader, count)
	profile.DataStrOutput = profile.decryptAES(encData)
	profile.DataStrOutput = append([]byte{RequestData}, profile.DataStrOutput...)
	switch err {
	case nil:
		log.Println(profile.DataStrOutput)
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

func (profile *Profile) verificationHandler(clientId uint, clientReader *bufio.Reader) {
	if clientId == -1 {
		return
	}
	profile.Friends[profile.getFriendNumber(clientId)].user.Ping = int(utils.GetUint16Reader(clientReader))
}

func (profile *Profile) networkHandler(clientReader *bufio.Reader) {
	var friendId uint
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

	signData := profile.decryptAES(signDataEncrypted)

	r := new(big.Int).SetBytes(signData[0 : signSize/2])
	s := new(big.Int).SetBytes(signData[signSize/2 : signSize/2])

	if profile.verifyData(metaDataEncrypted, *r, *s) == true {
		metaData := profile.decryptAES(metaDataEncrypted)
		username := metaData[0:userNameSize]
		fromUsername := metaData[userNameSize:fromUserNameSize]

		fmt.Println("UserNameSize:", userNameSize, " FromUserNameSize:", fromUserNameSize, " Username:", username,
			" Depth:", requestDepth, " BackTrace:", backTrace)
		if profile.Username == string(username) {
			//TODO: add friend request

			//friendId = addUser(User{profile: profile.Profile, username: string(username), publicKey: EncPublicKey(publicKey)})
			//profile.addFriendRequest(friendId, 1)

			fmt.Println("Friend Request done, Request from:", string(fromUsername), "Accept?")
			profile.DataStrOutput = append([]byte{RequestNetwork}, fromUsername...)
			profile.DataStrOutput = append(profile.DataStrOutput, publicKey...)
			profile.DataStrOutput = append(profile.DataStrOutput, backTrace...)

			UpdateUI(int(userNameSize), friendId)
			return
		}
	}

	requestDepth--
	//Required: Friends.ping && Friends.is_online
	if requestDepth > 0 {
		encrypted := make([]byte, 0)
		profile.buildEncryptedPart(&encrypted, publicKey, signData, metaDataEncrypted)
		profile.writeFindFriendRequestSecondary(int(requestDepth), int(requestDegree), friendId, backTrace, encrypted)
	}
}
