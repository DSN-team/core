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
}

type FriendRequestMeta struct {
	ToUsername, FromUsername string
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
	var request FriendRequest
	var requestMeta FriendRequestMeta

	requestEncoder := gob.NewEncoder(&requestBuffer)
	//build request meta
	requestMeta.ToUsername = user.Username
	requestMeta.FromUsername = cur.Username
	//encode request meta
	err = requestEncoder.Encode(requestMeta)
	ErrHandler(err)

	//encrypt meta
	encryptedMetaData := cur.encryptAES(user.PublicKey, requestBuffer.Bytes())

	profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)
	//build request meta
	request.FromPublicKey = profilePublicKey
	request.MetaDataEncrypted = encryptedMetaData

	request.Degree = 0
	request.Depth = 0
	cur.writeFindFriendRequestDirect(request, user)
	request.Degree = 2
	request.Depth = 2
	cur.writeFindFriendRequestSecondary(request, -1)
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
	log.Println("Add friend, ToUsername:", username, "address:", address, "publicKey:", publicKey)
	user := cur.getUserByUsername(username)
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

func (cur *Profile) LoadFriendsRequestsIn() int {
	cur.FriendRequestsIn = cur.GetFriendRequestsIn()
	return len(cur.FriendRequestsIn)
}

func (cur *Profile) LoadFriendsRequestsOut() int {
	cur.FriendRequestsOut = cur.GetFriendRequestsOut()
	return len(cur.FriendRequestsIn)
}
