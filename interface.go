package core

import (
	"fmt"
)

// UpdateUI size, id
var UpdateUI = func(int, uint) {}

func (profile *Profile) Register(username, password, address string) bool {
	profile.PrivateKey = genProfileKey()
	if profile.PrivateKey == nil {
		return false
	}
	profile.Username = username
	profile.Password = password
	profile.Address = address
	profile.addProfile()
	return true
}

func (profile *Profile) Login(password string, pos int) (result bool) {
	getProfile(Profiles[pos].ID, profile)
	if profile == nil {
		return false
	}
	result = profile.decProfileKey([]byte(profile.PrivateKeyString), password)
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
