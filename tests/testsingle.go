package main

import (
	"fmt"
	"github.com/DSN-team/core"
	"github.com/DSN-team/core/tests/utils"
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
func singleConnection(from, to *core.Profile) {
	from.AddFriend(to.Username, to.Address, to.GetProfilePublicKey())
	from.LoadFriends()
	from.ConnectToFriends()
	from.DataStrInput.Io = make([]byte, 128)
	from.DataStrOutput.Io = make([]byte, 128)
}

func main() {
	core.StartDB()
	core.LoadProfiles()
	profile0 := runProfile("0")
	profile1 := runProfile("1")
	go singleConnection(profile0, profile1)
	go singleConnection(profile1, profile0)
	delayedCall(profile0, profile1, "test")

	//Hold main thread
	for {
		time.Sleep(10)
	}
}

func delayedCall(from, to *core.Profile, msg string) {
	go func() {
		time.Sleep(450 * time.Millisecond)
		fmt.Println(utils.ConnectionsToString(from))
		for i := 0; i < len(msg); i++ {
			from.DataStrInput.Io[i] = msg[i]
		}
		from.WriteBytes(from.Friends[0].Id, len(msg))
		time.Sleep(400 * time.Millisecond)
		log.Println("got it:", to.DataStrOutput.Io)
		log.Println("got it as string:", string(to.DataStrOutput.Io))
	}()
}
