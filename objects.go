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
	IsFriend  bool
}

type Profile struct {
	Id          int
	Username    string
	Password    string
	Address     string
	PrivateKey  *ecdsa.PrivateKey
	Friends     []User
	connections sync.Map
}

type ShowProfile struct {
	Id       int
	Username string
}
