package core

import (
	"crypto/ecdsa"
	"sync"
)

type User struct {
	Id        int
	Username  string
	Address   string
	PublicKey *ecdsa.PublicKey
	Ping      int
	IsOnline  bool
	IsFriend  bool
}

type Profile struct {
	Password   string
	ThisUser   User
	PrivateKey *ecdsa.PrivateKey
	//From ID to index
	FriendsIDXs    sync.Map
	Connections    sync.Map
	Friends        []User
	FriendRequests []FriendRequest
	DataStrOutput  []byte
	DataStrInput   []byte
}

type ShowProfile struct {
	Id       int
	Username string
}

type FriendRequest struct {
	PublicKey string
	Username  string
}
