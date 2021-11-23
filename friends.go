package core

import (
	"github.com/DSN-team/core/utils"
	"log"
	"sort"
	"time"
)

func (profile *Profile) sortFriends() {
	sort.Slice(profile.Friends, func(i, j int) bool {
		return profile.Friends[i].user.Ping < profile.Friends[j].user.Ping
	})
}

func (profile *Profile) getFriendNumber(id uint) int {
	output, _ := profile.FriendsIDXs.Load(id)
	return output.(int)
}

func (profile *Profile) WriteFindFriendRequest(user *User) {
	log.Println("Write Friend Request, Friend:", user.Username)
	profilePublicKey := MarshalPublicKey(&profile.PrivateKey.PublicKey)

	metadata := make([]byte, 0)
	utils.SetBytes(&metadata, []byte(user.Username))
	utils.SetBytes(&metadata, []byte(profile.Username))

	encryptedData := profile.encryptAES(user.PublicKey, metadata)

	sign := make([]byte, 0)
	r, s := profile.signData(encryptedData)

	utils.SetBytes(&sign, profilePublicKey)
	utils.SetBytes(&sign, r.Bytes())
	utils.SetBytes(&sign, s.Bytes())

	encryptedSign := profile.encryptAES(user.PublicKey, sign)

	request := make([]byte, 0)
	utils.SetUint16(&request, uint16(len(user.Username)))
	utils.SetUint16(&request, uint16(len(profile.Username)))

	profile.buildEncryptedPart(&request, profilePublicKey, encryptedData, encryptedSign)

	profile.writeFindFriendRequestDirect(0, 0, user, []byte{}, request)
	profile.writeFindFriendRequestSecondary(2, 2, 0, []byte{}, request)
}

func (profile *Profile) buildEncryptedPart(request *[]byte, key, sign, data []byte) {
	log.Println("Building encrypted part")
	utils.SetBytes(request, key)
	utils.SetUint32(request, uint32(len(data)))
	utils.SetBytes(request, data)
	utils.SetUint32(request, uint32(len(sign)))
	utils.SetBytes(request, sign)
}

func (profile *Profile) writeFindFriendRequestSecondary(depth, degree int, fromID int, previousTrace []byte, encrypted []byte) {
	log.Print("Write find Friend Request secondary, depth:", depth, "degree:", degree, "fromID:", fromID)
	for i := 0; i < len(profile.Friends); i++ {
		if i >= depth {
			break
		}
		sendTo := &profile.Friends[i]
		if sendTo.ID == uint(fromID) {
			continue
		}
		go func(sendTo *Friend) {
			request := make([]byte, 0)
			utils.SetUint8(&request, RequestNetwork)
			utils.SetUint8(&request, uint8(depth))
			utils.SetUint8(&request, uint8(degree))
			newTrace := make([]byte, 0)
			utils.SetBytes(&newTrace, previousTrace)
			utils.SetUint8(&newTrace, uint8(sendTo.ID))
			utils.SetUint8(&request, uint8(len(newTrace))) //BackTraceSize
			utils.SetBytes(&request, newTrace)             //BackTrace
			utils.SetBytes(&request, encrypted)
			profile.WriteRequest(sendTo.ID, request)
		}(sendTo)
	}
}

func (profile *Profile) writeFindFriendRequestDirect(depth, degree int, sendTo *User, previousTrace []byte, encrypted []byte) {
	log.Print("Write find Friend Request direct, depth:", depth, "degree:", degree)
	request := make([]byte, 0)
	utils.SetUint8(&request, RequestNetwork)
	utils.SetUint8(&request, uint8(depth))
	utils.SetUint8(&request, uint8(degree))
	newTrace := make([]byte, 0)
	utils.SetBytes(&newTrace, previousTrace)
	utils.SetUint8(&newTrace, uint8(sendTo.ID))
	utils.SetUint8(&request, uint8(len(newTrace))) //BackTraceSize
	utils.SetBytes(&request, newTrace)             //BackTrace
	utils.SetBytes(&request, encrypted)
	profile.connect(sendTo)
	profile.WriteRequest(sendTo.ID, request)
}

func (profile *Profile) AddFriend(username, address, publicKey string) {
	log.Println("Add Friend, username:", username, "address:", address, "publicKey:", publicKey)
	user := profile.getUserByPublicKey(publicKey)
	if user == nil {
		user.profile = profile
		user.Username = username
		user.Address = address
		user.PublicKeyString = publicKey
		user.ID = addUser(user)
	}
	//profile.Friends = append(profile.Friends, User)
	//profile.FriendsIDXs.Store(len(profile.Friends)-1, profile.Friends[len(profile.Friends)-1])
	//TODO: send Request to target
	//addRequest(user.ID, 0)
	go profile.WriteFindFriendRequest(user)
	//else {
	//	profile.editUser(id, User)
	//}
}

func (profile *Profile) LoadFriends() int {
	log.Println("Loading Friends from db")
	//TODO: Load friends
	//profile.Friends = profile.getFriends()
	for i := 0; i < len(profile.Friends); i++ {
		profile.FriendsIDXs.Store(profile.Friends[i].user.ID, i)
	}
	return len(profile.Friends)
}

func (profile *Profile) ConnectToFriends() {
	go func() {
		for i := 0; i < len(profile.Friends); i++ {
			log.Println("Connecting to Friend:", profile.Friends[i].user.Username)
			go profile.connect(&profile.Friends[i].user)
		}
		for {
			time.Sleep(250 * time.Millisecond)
			for i := 0; i < len(profile.Friends); i++ {
				profile.connect(&profile.Friends[i].user)
			}
		}
	}()
}

func (profile *Profile) ConnectToFriend(pos uint) {
	go profile.connect(&profile.Friends[profile.getFriendNumber(pos)].user)
}

func (profile *Profile) LoadFriendsRequests() int {
	//TODO: Load friends requests
	//profile.Requests = profile.GetFriendRequests()
	return len(profile.Requests)
}
