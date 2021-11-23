package core

import (
	"fmt"
	"log"
)

// UpdateUI size, id
var UpdateUI = func(int, int) {}

func (cur *Profile) Register(username, password, address string) bool {
	key := genProfileKey()
	if key == nil {
		return false
	}
	cur.User.Username, cur.Password, cur.User.Address, cur.PrivateKey = username, password, address, key
	log.Println(cur)
	cur.User.Id = addProfile(cur)
	return true
}

func (cur *Profile) Login(password string, pos int) (result bool) {
	var privateKeyEncBytes []byte
	cur.User.Id = Profiles[pos].Id
	cur.User.Username, cur.User.Address, privateKeyEncBytes = getProfileByID(Profiles[pos].Id)
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
