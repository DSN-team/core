package core

import (
	"crypto/ecdsa"
	"gorm.io/gorm"
	"sync"
)

type Profile struct {
	gorm.Model
	Username         string
	Address          string
	PrivateKeyString string
	Password         string            `gorm:"-"`
	PrivateKey       *ecdsa.PrivateKey `gorm:"-"`
	FriendsIDXs      sync.Map          `gorm:"-"`
	Connections      sync.Map          `gorm:"-"`
	Friends          []Friend          `gorm:"-"`
	Requests         []Request         `gorm:"-"`
	DataStrOutput    []byte            `gorm:"-"`
	DataStrInput     []byte            `gorm:"-"`
}

type User struct {
	gorm.Model
	profile         *Profile
	Username        string
	Address         string
	PublicKeyString string
	PublicKey       *ecdsa.PublicKey `gorm:"-"`
	Ping            int              `gorm:"-"`
	IsOnline        bool             `gorm:"-"`
	IsFriend        bool             `gorm:"-"`
}

type Friend struct {
	gorm.Model
	user User
}

type Request struct {
	gorm.Model
	profile   Profile
	User      User
	direction uint
}

type ShowProfile struct {
	Id       uint
	Username string
}
