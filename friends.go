package core

import (
	"crypto/ecdsa"
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

func (cur *Profile) WriteFindFriendRequest(friendUsername string, key *ecdsa.PublicKey) {
	profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)

	metadata := make([]byte, 0)
	utils.SetBytes(&metadata, []byte(friendUsername))
	utils.SetBytes(&metadata, []byte(cur.ThisUser.Username))

	encryptedData := cur.encryptAES(key, metadata)

	sign := make([]byte, 0)
	r, s := cur.signData(encryptedData)

	utils.SetBytes(&sign, profilePublicKey)
	utils.SetBytes(&sign, r.Bytes())
	utils.SetBytes(&sign, s.Bytes())

	encryptedSign := cur.encryptAES(key, sign)

	request := make([]byte, 0)
	utils.SetUint16(&request, uint16(len(friendUsername)))
	utils.SetUint16(&request, uint16(len(cur.ThisUser.Username)))

	cur.buildEncryptedPart(&request, profilePublicKey, encryptedData, encryptedSign)

	cur.writeFindFriendRequestSecondary(2, 2, -1, []byte{}, request)
}
func (cur *Profile) buildEncryptedPart(request *[]byte, key, sign, data []byte) {
	utils.SetBytes(request, key)
	utils.SetUint32(request, uint32(len(data)))
	utils.SetBytes(request, data)
	utils.SetUint32(request, uint32(len(sign)))
	utils.SetBytes(request, sign)
}

func (cur *Profile) writeFindFriendRequestSecondary(depth, degree, fromID int, previousTrace []byte, encrypted []byte) {

	for i := 0; i < len(cur.Friends); i++ {
		if i >= depth {
			break
		}
		sendTo := cur.Friends[i]
		if sendTo.Id == fromID {
			continue
		}
		request := make([]byte, 0)
		utils.SetUint8(&request, RequestNetwork)
		utils.SetUint8(&request, uint8(depth))
		utils.SetUint8(&request, uint8(degree))
		newTrace := make([]byte, 0)
		utils.SetBytes(&newTrace, previousTrace)
		utils.SetUint8(&newTrace, uint8(sendTo.Id))
		utils.SetUint8(&request, uint8(len(newTrace))) //BackTraceSize
		utils.SetBytes(&request, newTrace)             //BackTrace
		utils.SetBytes(&request, encrypted)
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
