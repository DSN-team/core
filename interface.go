package core

import (
	"fmt"
	"log"
)

// UpdateUI size, id
var UpdateUI = func(int, int) {}

func (cur *Profile) Register(username, password, address string) bool {
	privateKey := genProfileKey()
	if privateKey == nil {
		return false
	}
	cur.Username, cur.Password, cur.Address, cur.PrivateKey = username, password, address, privateKey
	log.Println(cur)
	addProfile(cur)
	return true
}

func (cur *Profile) Login(password string, pos int) (result bool) {
	var privateKeyEncBytes []byte
	profile := getProfileByID(Profiles[pos].ID)
	cur.ID = profile.ID
	cur.Username = profile.Username
	cur.Address = profile.Address
	cur.PrivateKeyString = profile.PrivateKeyString

	fmt.Println("privateKeyEncBytes: ", privateKeyEncBytes)
	if privateKeyEncBytes == nil {
		return false
	}
	result = cur.decProfileKey(privateKeyEncBytes, password)
	fmt.Println("Login status:", result)
	return
}

func UsernamePos(username string) int {
	profiles := getProfiles()
	pos := -1
	for i := 0; i < len(profiles); i++ {
		if profiles[i].Username == username {
			pos = i
			break
		}
	}
	return pos
}
