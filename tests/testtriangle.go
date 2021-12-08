package main

import (
	"github.com/DSN-team/core/tests/utils"
	"time"
)

func main() {
	utils.InitTest(true)
	profile0 := utils.RunProfile("0")
	profile1 := utils.RunProfile("1")
	profile2 := utils.RunProfile("2")
	utils.CreateNetwork(profile0, profile1)
	time.Sleep(100 * time.Millisecond)
	utils.CreateNetwork(profile1, profile0)
	time.Sleep(100 * time.Millisecond)
	utils.CreateNetwork(profile2, profile1)

	//fmt.Println("requests:", profile0.getFriendRequestsIn())
	go utils.StartConnection(profile0)
	go utils.StartConnection(profile1)
	go utils.StartConnection(profile2)
	profile0.LoadFriendsRequestsIn()
	profile1.LoadFriendsRequestsIn()
	profile2.LoadFriendsRequestsIn()
	for i := 0; i < len(profile0.FriendRequestsIn); i++ {
		profile0.AcceptFriendRequest(&profile0.FriendRequestsIn[i])
		profile0.LoadFriends()
		profile0.AnswerFindFriendRequest(profile0.FriendRequestsIn[i])
		profile1.AcceptFriendRequest(&profile1.FriendRequestsIn[i])
		profile1.LoadFriends()
		profile1.AnswerFindFriendRequest(profile1.FriendRequestsIn[i])
		profile2.AcceptFriendRequest(&profile2.FriendRequestsIn[i])
		profile2.LoadFriends()
		profile2.AnswerFindFriendRequest(profile2.FriendRequestsIn[i])
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)

	utils.CreateNetwork(profile0, profile2)
	profile0.LoadFriendsRequestsIn()
	for i := 0; i < len(profile0.FriendRequestsIn); i++ {
		profile0.AcceptFriendRequest(&profile0.FriendRequestsIn[i])
		profile0.LoadFriends()
		profile0.AnswerFindFriendRequest(profile0.FriendRequestsIn[i])
	}

	time.Sleep(200 * time.Millisecond)
	utils.DelayedCall(profile2, profile0, "test")

	//Hold main thread
	for {
		time.Sleep(10)
	}
}
