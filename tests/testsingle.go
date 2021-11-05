package main

import (
	"github.com/DSN-team/core"
)

func runProfile(nameNumber string) core.Profile {
	username := "Node:" + nameNumber
	password := "Pass:" + nameNumber
	println("node:" + nameNumber)
	println("username:" + username)
	println("password:" + password)
	pos := core.UsernamePos(username)
	singleProfile := core.Profile{}
	if pos == -1 {
		singleProfile.Register(username, password) //already logged in after register
	} else {
		singleProfile.Login(password, pos)
	}
	singleProfile.LoadFriends()
	//singleProfile.RunServer("127.0.0.0:"+strconv.FormatInt(int64(port),10))
	return singleProfile
}

func main() {
	core.StartDB()
	core.LoadProfiles()
	profile0 := runProfile("0")
	profile1 := runProfile("1")

	profile0.AddFriend(profile1.Address, profile1.GetProfilePublicKey())
	profile1.AddFriend(profile0.Address, profile0.GetProfilePublicKey())
	profile0.ConnectToFriends()
	profile1.ConnectToFriends()
}
