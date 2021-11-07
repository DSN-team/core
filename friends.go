package core

import (
	"encoding/binary"
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

func (cur *Profile) FindFriendRequest(username string) {
	cur.FindFriendRequestSecondary(username, 2, 2)
}

func (cur *Profile) FindFriendRequestSecondary(username string, depth, degree int) {
	request := make([]byte, 8)
	binary.BigEndian.PutUint64(request, uint64(len(username)))
	//Required sort friends
}

func (cur *Profile) AddFriend(username, address, publicKey string) {
	decryptedPublicKey := UnmarshalPublicKey(DecPublicKey(publicKey))
	id := cur.searchUser(username)
	user := User{Username: username, Address: address, PublicKey: &decryptedPublicKey, IsFriend: true}
	if id == -1 {
		cur.addUser(user)
		cur.FriendsIDXs.Store(cur.Friends[len(cur.Friends)-1].Id, len(cur.Friends)-1)
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

func (cur *Profile) ConnectToFriend(userId int) {
	go cur.connect(cur.getFriendNumber(userId))
}
