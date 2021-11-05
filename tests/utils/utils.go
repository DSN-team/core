package utils

import (
	"github.com/DSN-team/core"
	"strconv"
)

func ProfileToString(user *core.Profile) (output string) {
	output += "Username:" + user.Username
	output += "Password:" + user.Password
	output += "Address:" + user.Address
	return output
}
func ConnectionsToString(profile *core.Profile) (output string) {
	for i := 0; i < len(profile.Friends); i++ {
		_, temp := profile.Connections.Load(profile.Friends[i].Id)
		output += "pos:" + strconv.FormatInt(int64(profile.Friends[i].Id), 10) + ":" + strconv.FormatBool(temp) + "\n"
	}
	return output
}
