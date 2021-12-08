package main

import (
	"github.com/DSN-team/core/tests/utils"
	"time"
)

func main() {
	utils.InitTest(true)
	profile0 := utils.RunProfile("0")
	profile1 := utils.RunProfile("1")
	utils.CreateNetwork(profile0, profile1)
	time.Sleep(100 * time.Millisecond)
	utils.CreateNetwork(profile1, profile0)

	//fmt.Println("requests:", profile0.getFriendRequestsIn())
	go utils.StartConnection(profile0)
	go utils.StartConnection(profile1)
	time.Sleep(100 * time.Millisecond)
	profile0.LoadFriendsRequestsIn()
	profile1.LoadFriendsRequestsIn()
	for i := 0; i < len(profile0.FriendRequestsIn); i++ {
		profile0.AcceptFriendRequest(&profile0.FriendRequestsIn[i])
		profile0.LoadFriends()
		profile0.AnswerFindFriendRequest(profile0.FriendRequestsIn[i])
	}
	time.Sleep(200 * time.Millisecond)
	utils.DelayedCall(profile1, profile0, "test")
	//Hold main thread
	for {
		time.Sleep(10)
	}
}
