package main

import "crypto/ecdsa"

type User struct {
	username string
	address  string
}

type Profile struct {
	username   string
	password   string
	address    string
	privateKey *ecdsa.PrivateKey
}

type ShowProfile struct {
	id       int
	username string
}
