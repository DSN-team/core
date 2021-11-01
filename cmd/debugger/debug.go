package main

import (
	"github.com/DSN-team/core"
	"os"
	"strconv"
)

func main() {
	args := os.Args
	username := "Node:" + args[1]
	password := "Pass:" + args[1]
	node := args[1]
	println("node:" + node)
	println("username:" + username)
	println("password:" + password)
	core.StartDB()
	core.Register(username, password) //already logged in after register
	//core.LoadProfiles()
	//core.Login(password, core.UsernamePos(username))
	//core.LoadFriends()
	port, _ := strconv.Atoi(node)
	port += 25
	switch node {
	case "0":
		{
			break
		}
	case "1":
		{

			break
		}
	case "2":
		{

			break
		}
	case "3":
		{

			break
		}
	case "4":
		{

			break
		}
	case "5":
		{

			break
		}
	case "6":
		{

			break
		}
	case "7":
		{

			break
		}
	}

}
