package core

import (
	"bytes"
	"encoding/gob"
	"github.com/DSN-team/core/utils"
	"log"
	"sort"
	"time"
)

type FriendRequest struct {
	Depth, Degree     int
	BackTrace         []byte
	FromPublicKey     []byte
	MetaDataEncrypted []byte
	SignEncrypted     []byte
}

type FriendRequestMeta struct {
	Username, FromUsername string
}
type FriendRequestSign struct {
	FromPublicKey []byte
	SignR, SignS  int64
}

func (cur *Profile) sortFriends() {
	sort.Slice(cur.Friends, func(i, j int) bool {
		return cur.Friends[i].Ping < cur.Friends[j].Ping
	})
}

func (cur *Profile) getFriendNumber(id int) int {
	output, _ := cur.FriendsIDXs.Load(id)
	return output.(int)
}

func (cur *Profile) WriteFindFriendRequest(user User) {
	log.Println("Write friend request, friend:", user.Username)
	var requestBuffer bytes.Buffer
	var signBuffer bytes.Buffer

	var request FriendRequest
	var requestMeta FriendRequestMeta
	var requestSign FriendRequestSign

	requestEncoder := gob.NewEncoder(&requestBuffer)
	signEncoder := gob.NewEncoder(&signBuffer)

	profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)

	//build request meta
	requestMeta.Username = user.Username
	requestMeta.FromUsername = cur.Username
	//encode request meta
	err = requestEncoder.Encode(requestMeta)
	ErrHandler(err)

	//encrypt meta
	encryptedMetaData := cur.encryptAES(user.PublicKey, requestBuffer.Bytes())

	r, s := cur.signData(encryptedMetaData)
	//build request sign
	requestSign.FromPublicKey = profilePublicKey
	requestSign.SignR = r.Int64()
	requestSign.SignS = s.Int64()
	//encode request sign
	err = signEncoder.Encode(requestSign)
	ErrHandler(err)
	//encrypt request sign
	encryptedSign := cur.encryptAES(user.PublicKey, signBuffer.Bytes())
	//build request meta and sign
	request.FromPublicKey = MarshalPublicKey(&cur.PrivateKey.PublicKey)
	request.MetaDataEncrypted = encryptedMetaData
	request.SignEncrypted = encryptedSign

	request.Degree = 0
	request.Depth = 0
	cur.writeFindFriendRequestDirect(request, user)
	request.Degree = 2
	request.Depth = 2
	cur.writeFindFriendRequestSecondary(request, -1)
}

func (cur *Profile) buildEncryptedPart(request *[]byte, key, sign, data []byte) {
	log.Println("Building encrypted part")
	utils.SetBytes(request, key)
	utils.SetUint32(request, uint32(len(data)))
	utils.SetBytes(request, data)
	utils.SetUint32(request, uint32(len(sign)))
	utils.SetBytes(request, sign)
}

func (cur *Profile) writeFindFriendRequestSecondary(request FriendRequest, fromID int) {
	for i := 0; i < len(cur.Friends); i++ {
		if i >= request.Depth {
			break
		}
		sendTo := cur.Friends[i]
		if int(sendTo.ID) == fromID {
			continue
		}
		go func(sendTo User) {
			cur.writeFindFriendRequestDirect(request, sendTo)
		}(sendTo)
	}
}

func (cur *Profile) writeFindFriendRequestDirect(friendRequest FriendRequest, sendTo User) {
	//log.Print("Write find friend friendRequest direct, depth:", depth, "degree:", degree)
	newTrace := make([]byte, 0)
	utils.SetBytes(&newTrace, friendRequest.BackTrace)
	utils.SetUint8(&newTrace, uint8(sendTo.ID))
	friendRequest.BackTrace = newTrace

	var friendRequestBuffer bytes.Buffer
	friendRequestEncoder := gob.NewEncoder(&friendRequestBuffer)
	err = friendRequestEncoder.Encode(friendRequest)
	ErrHandler(err)

	request := Request{RequestType: utils.RequestNetwork, PublicKey: MarshalPublicKey(&cur.PrivateKey.PublicKey), Data: friendRequestBuffer.Bytes()}
	cur.WriteRequest(sendTo, request)
}
func (cur *Profile) AddFriend(username, address, publicKey string) {
	log.Println("Add friend, Username:", username, "address:", address, "publicKey:", publicKey)
	user := cur.searchUser(username)
	if user.ID == 0 {
		user = User{ProfileID: cur.ID, Username: username, Address: address, PublicKeyString: publicKey, IsFriend: false}
		cur.addUser(&user)
		cur.Friends = append(cur.Friends, user)
		cur.FriendsIDXs.Store(user.ID, len(cur.Friends)-1)
	}
	decryptedPublicKey := UnmarshalPublicKey(DecodeKey(user.PublicKeyString))
	user.PublicKey = &decryptedPublicKey
	cur.addFriendRequest(user.ID, 0)
	go cur.WriteFindFriendRequest(user)
}

func (cur *Profile) LoadFriends() int {
	log.Println("Loading Friends from db")
	cur.Friends = cur.getFriends()
	for i := 0; i < len(cur.Friends); i++ {
		publicKey := UnmarshalPublicKey(DecodeKey(cur.Friends[i].PublicKeyString))
		cur.Friends[i].PublicKey = &publicKey
		cur.FriendsIDXs.Store(cur.Friends[i].ID, i)
	}
	return len(cur.Friends)
}

func (cur *Profile) ConnectToFriends() {
	go func() {
		for i := 0; i < len(cur.Friends); i++ {
			log.Println("Connecting to friend:", cur.Friends[i].Username, "number:", i)
			go cur.connect(cur.Friends[i])
		}
		for {
			time.Sleep(250 * time.Millisecond)
			for i := 0; i < len(cur.Friends); i++ {
				cur.connect(cur.Friends[i])
			}
		}
	}()

}

func (cur *Profile) ConnectToFriend(pos int) {
	go cur.connect(cur.Friends[cur.getFriendNumber(pos)])
}

func (cur *Profile) LoadFriendsRequests() int {
	cur.FriendRequests = cur.GetFriendRequests()
	return len(cur.FriendRequests)
}
