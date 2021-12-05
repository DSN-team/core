package core

import (
	"crypto/ecdsa"
	"gorm.io/gorm"
	"sync"
)

type User struct {
	gorm.Model
	ProfileID       uint
	Username        string
	Address         string
	PublicKeyString string
	IsFriend        bool

	PublicKey *ecdsa.PublicKey `gorm:"-"`
	Ping      int              `gorm:"-"`
	IsOnline  bool             `gorm:"-"`

	UserRequest []UserRequest
}

type Profile struct {
	gorm.Model
	Username         string
	Address          string
	Password         string
	PrivateKeyString string

	Friends           []User
	FriendRequestsIn  []UserRequest
	FriendRequestsOut []UserRequest

	FriendsIDXs   sync.Map          `gorm:"-"`
	Connections   sync.Map          `gorm:"-"`
	DataStrOutput []byte            `gorm:"-"`
	DataStrInput  []byte            `gorm:"-"`
	PrivateKey    *ecdsa.PrivateKey `gorm:"-"`
}

type ShowProfile struct {
	ID       uint
	Username string
}

type UserRequest struct {
	gorm.Model
	ProfileID uint
	UserID    uint
	Direction int
	Status    int
}
