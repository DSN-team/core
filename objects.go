package core

import (
	"crypto/ecdsa"
	"sync"
)

type strBuffer struct {
	Io []byte
}

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
	Password string
	thisUser User
	//Deprecated: Use
	Id int
	//Deprecated: Use
	Username string
	//Deprecated: Use
	Address    string
	PrivateKey *ecdsa.PrivateKey
	//From ID to index
	FriendsIDXs   sync.Map
	Friends       []User
	Connections   sync.Map
	DataStrOutput strBuffer
	DataStrInput  strBuffer
}

type ShowProfile struct {
	Id       int
	Username string
}
