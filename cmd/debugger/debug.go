package main

import (
	"github.com/DSN-team/core"
	"os"
)

func main() {
	args := os.Args
	username := args[0]
	password := args[1]
	node := args[2]
	core.Register(username, password)
	core.Login(password, core.UsernamePos(username))
	core.LoadProfiles()
	core.LoadFriends()
	switch node {
	case "0":
		{

			break
		}
	case "1":
		{

			break
		}
	}

}
