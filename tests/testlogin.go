package main

import "github.com/DSN-team/core"

func main() {
	utils.InitTest()
	result := SelectedProfile.Login("test1", 1)
	println(result)
}
