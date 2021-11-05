package utils

import (
	"github.com/DSN-team/core"
	"strconv"
)

func RunProfile(nameNumber string) *core.Profile {
	username := "Node:" + nameNumber
	password := "Pass:" + nameNumber
	println("node:" + nameNumber)
	println("username:" + username)
	println("password:" + password)
	pos := core.UsernamePos(username)
	singleProfile := &core.Profile{}
	if pos == -1 {
		singleProfile.Register(username, password, "127.0.0.1:3"+nameNumber) //already logged in after register
	} else {
		singleProfile.Login(password, "127.0.0.1:3"+nameNumber, pos)
		singleProfile.LoadFriends()
	}
	singleProfile.RunServer("127.0.0.1:3" + nameNumber)
	return singleProfile
}

func CreateNetwork(from, to *core.Profile) {
	from.AddFriend(to.Username, to.Address, to.GetProfilePublicKey())
}

func StartConnection(from *core.Profile) {
	if from.DataStrInput.Io == nil {
		from.DataStrInput.Io = make([]byte, 128)
	}
	if from.DataStrOutput.Io == nil {
		from.DataStrOutput.Io = make([]byte, 128)
	}
	from.LoadFriends()
	from.ConnectToFriends()
}

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
