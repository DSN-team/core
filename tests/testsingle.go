package main

import (
	"github.com/DSN-team/core"
	"log"
	"time"
)

func runProfile(nameNumber string) *core.Profile {
	username := "Node:" + nameNumber
	password := "Pass:" + nameNumber
	println("node:" + nameNumber)
	println("username:" + username)
	println("password:" + password)
	pos := core.UsernamePos(username)
	singleProfile := &core.Profile{}
	if pos == -1 {
		singleProfile.Register(username, password, "127.0.0.1:2"+nameNumber) //already logged in after register
	} else {
		singleProfile.Login(password, "127.0.0.1:2"+nameNumber, pos)
	}
	singleProfile.LoadFriends()
	singleProfile.RunServer("127.0.0.1:2" + nameNumber)
	return singleProfile
}

func main() {
	core.StartDB()
	core.LoadProfiles()
	profile0 := runProfile("0")
	profile1 := runProfile("1")

	profile0.AddFriend(profile1.Username, profile1.Address, profile1.GetProfilePublicKey())
	log.Println("friends count: ", profile0.LoadFriends())
	profile1.AddFriend(profile0.Username, profile0.Address, profile0.GetProfilePublicKey())
	profile1.LoadFriends()
	go profile0.ConnectToFriends()
	go profile1.ConnectToFriends()
	profile0.DataStrInput.Io = make([]byte, 10)
	profile1.DataStrInput.Io = make([]byte, 10)
	profile0.DataStrOutput.Io = make([]byte, 10)
	profile1.DataStrOutput.Io = make([]byte, 10)
	delayedCall(profile0, profile1)

	//Hold main thread
	for {
		time.Sleep(10)
	}
}

func delayedCall(from, to *core.Profile) {
	go func() {
		time.Sleep(450 * time.Millisecond)
		for i := 0; i < 10; i++ {
			from.DataStrInput.Io[i] = 10
		}
		from.WriteBytes(to.Id, 10)
		time.Sleep(400 * time.Millisecond)
		log.Println("got it:", to.DataStrOutput.Io)
	}()
}
