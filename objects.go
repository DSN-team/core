package core

import (
	"crypto/ecdsa"
	"gorm.io/gorm"
	"sync"
)

type User struct {
	gorm.Model
	//ID             uint   `gorm:"primaryKey; not null"`
	Id           int
	Username     string
	Address      string
	PublicKey    *ecdsa.PublicKey `gorm:"-"`
	PublicKeyStr string
	Ping         int
	IsOnline     bool
	IsFriend     bool
}

type Profile struct {
	gorm.Model
	ID       uint `gorm:"primaryKey; not null"`
	Password string
	//Username string
	Address    string
	User       User              `gorm:"foreignKey:ID"`
	PrivateKey *ecdsa.PrivateKey `gorm:"-"`
	//From Id to index
	FriendsIDXs    sync.Map        `gorm:"-"`
	Connections    sync.Map        `gorm:"-"`
	Friends        []User          `gorm:"-"`
	FriendRequests []FriendRequest `gorm:"-"`
	DataStrOutput  []byte          `gorm:"-"`
	DataStrInput   []byte          `gorm:"-"`
}
type NameProf struct {
	User ShowProfile `gorm:"foreignKey:Id"`
}

type ShowProfile struct {
	Id       int
	Username string
}

type FriendRequest struct {
	gorm.Model
	ID        uint `gorm:"primaryKey; not null"`
	PublicKey string
	Username  string
}
