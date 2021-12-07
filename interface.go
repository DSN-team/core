package core

import (
	"fmt"
	"log"
)

// UpdateUI size, id
var UpdateUI = func(int, int) {}

func (cur *Profile) Register(username, password, address string) bool {
	fmt.Println("Register", username, password, address)
	privateKey := genProfileKey()
	if privateKey == nil {
		return false
	}
	cur.Username, cur.Password, cur.Address, cur.PrivateKey = username, password, address, privateKey
	cur.encProfileKey()
	log.Println(cur)
	addProfile(cur)
	return true
}

func (cur *Profile) Login(password string, pos int) (result bool) {
	fmt.Println("Login", password, pos)
	profile := getProfileByID(Profiles[pos].ID)
	cur.Password = password
	cur.ID, cur.Username, cur.Address, cur.PrivateKeyString = profile.ID, profile.Username, profile.Address, profile.PrivateKeyString
	result = cur.decProfileKey(cur.PrivateKeyString, password)
	fmt.Println("Login status:", result)
	return
}

func UsernamePos(username string) int {
	fmt.Println("Getting username pos")
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
