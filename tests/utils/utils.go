package utils

import (
	"github.com/DSN-team/core"
	"os"
	"strconv"
	"time"
)

const baseAddress = "127.0.0.1:300"

func InitTest(clear bool) {
	//Sleep between retrying
	if clear {
		path := "data.db"
		if _, err := os.Stat(path); err == nil {
			err = os.Remove(path)
		}
	}
	time.Sleep(250 * time.Millisecond)
	core.StartDB()
	core.LoadProfiles()
}

func RunProfile(nameNumber string) *core.Profile {
	username := "Node" + nameNumber
	password := "Pass" + nameNumber
	println("node:" + nameNumber)
	println("username:" + username)
	println("password:" + password)
	pos := core.UsernamePos(username)
	singleProfile := &core.Profile{}
	if pos == -1 {
		singleProfile.Register(username, password, baseAddress+nameNumber) //already logged in after register
	} else {
		singleProfile.Login(password, pos)
		singleProfile.LoadFriends()
	}
	singleProfile.RunServer(baseAddress + nameNumber)
	return singleProfile
}

func CreateNetwork(from, to *core.Profile) {
	from.AddFriend(to.Username, to.Address, to.GetProfilePublicKey())
}

func StartConnection(from *core.Profile) {
	if from.DataStrInput == nil {
		from.DataStrInput = make([]byte, 128)
	}
	if from.DataStrOutput == nil {
		from.DataStrOutput = make([]byte, 128)
	}
	from.LoadFriends()
	from.ConnectToFriends()
}

func ProfileToString(user *core.Profile) (output string) {
	output += "ToUsername:" + user.Username
	output += "Password:" + user.Password
	output += "Address:" + user.Address
	return output
}
func ConnectionsToString(profile *core.Profile) (output string) {
	for i := 0; i < len(profile.Friends); i++ {
		_, temp := profile.Connections.Load(profile.Friends[i].ID)
		output += "pos:" + strconv.FormatInt(int64(profile.Friends[i].ID), 10) + ":" + strconv.FormatBool(temp) + "\n"
	}
	return output
}
