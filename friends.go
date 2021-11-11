package core

import (
	"github.com/DSN-team/core/utils"
	"sort"
)

func (cur *Profile) sortFriends() {
	sort.Slice(cur.Friends, func(i, j int) bool {
		return cur.Friends[i].Ping < cur.Friends[j].Ping
	})
}

func (cur *Profile) getFriendNumber(id int) int {
	output, _ := cur.FriendsIDXs.Load(id)
	return output.(int)
}

func (cur *Profile) WriteFindFriendRequest(username, key string) {
	encrypt := make([]byte, 0)
	utils.SetBytes(&encrypt, []byte(username))
	utils.SetBytes(&encrypt, []byte(cur.ThisUser.Username))

	//key:= UnmarshalPublicKey(DecPublicKey(keystr))
	profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)
	utils.SetBytes(&encrypt, profilePublicKey)
	//TODO encryption
	//encrypted := cur.encryptAES(&key,encrypt)
	encrypted := encrypt
	cur.WriteFindFriendRequestSecondary(username, 2, 2, -1, encrypted)
}

func (cur *Profile) WriteFindFriendRequestSecondary(username string, depth, degree, fromID int, encrypted []byte) {
	request := make([]byte, 0)
	utils.SetUint8(&request, RequestNetwork)

	utils.SetUint16(&request, uint16(len(username)))
	utils.SetUint16(&request, uint16(len(cur.ThisUser.Username)))
	utils.SetUint8(&request, uint8(depth))
	utils.SetUint8(&request, uint8(degree))
	utils.SetUint8(&request, uint8(0)) //BackTraceSize

	utils.SetBytes(&request, encrypted)
	//utils.SetBytes(&request,[]byte{0})//BackTrace

	for i := 0; i < len(cur.Friends); i++ {
		if i >= depth {
			break
		}
		sendTo := cur.Friends[i]
		if sendTo.Id == fromID {
			continue
		}
		cur.WriteRequest(sendTo.Id, request)
	}
}

func (cur *Profile) AddFriend(username, address, publicKey string) {
	decryptedPublicKey := UnmarshalPublicKey(DecPublicKey(publicKey))
	id := cur.searchUser(username)
	user := User{Username: username, Address: address, PublicKey: &decryptedPublicKey, IsFriend: true}
	if id == -1 {
		user.Id = cur.addUser(user)
		cur.Friends = append(cur.Friends, user)
		cur.FriendsIDXs.Store(len(cur.Friends)-1, cur.Friends[len(cur.Friends)-1])
	} else {
		cur.editUser(id, user)
	}
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

func (cur *Profile) ConnectToFriend(pos int) {
	go cur.connect(cur.getFriendNumber(pos))
}
