package core

import (
	"github.com/DSN-team/core/utils"
	"log"
	"sort"
	"time"
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

func (cur *Profile) WriteFindFriendRequest(user User) {
	log.Println("Write friend request, friend:", user.Username)
	profilePublicKey := MarshalPublicKey(&cur.PrivateKey.PublicKey)

	metadata := make([]byte, 0)
	utils.SetBytes(&metadata, []byte(user.Username))
	utils.SetBytes(&metadata, []byte(cur.Username))

	encryptedData := cur.encryptAES(user.PublicKey, metadata)

	sign := make([]byte, 0)
	r, s := cur.signData(encryptedData)

	utils.SetBytes(&sign, profilePublicKey)
	utils.SetBytes(&sign, r.Bytes())
	utils.SetBytes(&sign, s.Bytes())

	encryptedSign := cur.encryptAES(user.PublicKey, sign)

	request := make([]byte, 0)
	utils.SetUint16(&request, uint16(len(user.Username)))
	utils.SetUint16(&request, uint16(len(cur.Username)))

	cur.buildEncryptedPart(&request, profilePublicKey, encryptedData, encryptedSign)

	cur.writeFindFriendRequestDirect(0, 0, user, []byte{}, request)
	cur.writeFindFriendRequestSecondary(2, 2, -1, []byte{}, request)
}

func (cur *Profile) buildEncryptedPart(request *[]byte, key, sign, data []byte) {
	log.Println("Building encrypted part")
	utils.SetBytes(request, key)
	utils.SetUint32(request, uint32(len(data)))
	utils.SetBytes(request, data)
	utils.SetUint32(request, uint32(len(sign)))
	utils.SetBytes(request, sign)
}

func (cur *Profile) writeFindFriendRequestSecondary(depth, degree, fromID int, previousTrace []byte, encrypted []byte) {
	log.Print("Write find friend request secondary, depth:", depth, "degree:", degree, "fromID:", fromID)
	for i := 0; i < len(cur.Friends); i++ {
		if i >= depth {
			break
		}
		sendTo := cur.Friends[i]
		if int(sendTo.ID) == fromID {
			continue
		}
		go func(sendTo User) {
			request := make([]byte, 0)
			utils.SetUint8(&request, utils.RequestNetwork)
			utils.SetUint8(&request, uint8(depth))
			utils.SetUint8(&request, uint8(degree))
			newTrace := make([]byte, 0)
			utils.SetBytes(&newTrace, previousTrace)
			utils.SetUint8(&newTrace, uint8(sendTo.ID))
			utils.SetUint8(&request, uint8(len(newTrace))) //BackTraceSize
			utils.SetBytes(&request, newTrace)             //BackTrace
			utils.SetBytes(&request, encrypted)
			cur.WriteRequest(int(sendTo.ID), request)
		}(sendTo)
	}
}

func (cur *Profile) writeFindFriendRequestDirect(depth, degree int, sendTo User, previousTrace []byte, encrypted []byte) {
	log.Print("Write find friend request direct, depth:", depth, "degree:", degree)
	request := make([]byte, 0)
	utils.SetUint8(&request, utils.RequestNetwork)
	utils.SetUint8(&request, uint8(depth))
	utils.SetUint8(&request, uint8(degree))
	newTrace := make([]byte, 0)
	utils.SetBytes(&newTrace, previousTrace)
	utils.SetUint8(&newTrace, uint8(sendTo.ID))
	utils.SetUint8(&request, uint8(len(newTrace))) //BackTraceSize
	utils.SetBytes(&request, newTrace)             //BackTrace
	utils.SetBytes(&request, encrypted)
	cur.connect(sendTo)
	cur.WriteRequest(int(sendTo.ID), request)
}

func (cur *Profile) AddFriend(username, address, publicKey string) {
	log.Println("Add friend, username:", username, "address:", address, "publicKey:", publicKey)
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
