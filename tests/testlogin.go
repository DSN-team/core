package main

import "github.com/DSN-team/core"

func main() {
	core.StartDB()
	SelectedProfile := core.Profile{}
	core.LoadProfiles()
	result := SelectedProfile.Login("test1", 1)
	println(result)
}
