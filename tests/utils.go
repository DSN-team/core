package main

import "github.com/DSN-team/core"

func profileToString(user core.Profile) (output string) {
	output += "Username:" + user.Username
	output += "Password:" + user.Password
	output += "Address:" + user.Address
	return output
}
