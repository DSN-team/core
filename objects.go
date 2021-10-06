package main

import "crypto/ecdsa"

type User struct {
	id        int
	username  string
	address   string
	publicKey *ecdsa.PublicKey
	isFriend  bool
}

type Profile struct {
	id         int
	username   string
	password   string
	address    string
	privateKey *ecdsa.PrivateKey
}

type ShowProfile struct {
	id       int
	username string
}
