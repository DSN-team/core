package main

import (
	"fmt"
	"github.com/DSN-team/core"
	"github.com/DSN-team/core/tests/utils"
	utils2 "github.com/DSN-team/core/utils"
	"log"
	"time"
)

func main() {
	core.UpdateUI = func(i int, client int) {
		log.Print("client:", client, "\n")
		//log.Print("got it:", to.DataStrOutput)
		//log.Println("got it as string:", string(to.DataStrOutput))
	}

	utils.InitTest()
	profile0 := utils.RunProfile("0")
	profile1 := utils.RunProfile("1")
	utils.CreateNetwork(profile0, profile1)
	time.Sleep(100 * time.Millisecond)
	utils.CreateNetwork(profile1, profile0)

	fmt.Println("requests:", profile0.GetFriendRequests())
	//go utils.StartConnection(profile0)
	//go utils.StartConnection(profile1)
	//time.Sleep(100 * time.Millisecond)

	//delayedCall(profile0, profile1, "test")
	//delayedCall(profile1, profile0, "test")

	//Hold main thread
	for {
		time.Sleep(10)
	}
}

func delayedCall(from, to *core.Profile, msg string) {
	for i := 0; i < len(msg); i++ {
		from.DataStrInput[i] = msg[i]
	}
	dataMessage := from.BuildDataMessage([]byte(msg), from.Friends[0].ID)
	request := core.Request{RequestType: utils2.RequestData, PublicKey: core.MarshalPublicKey(&from.PrivateKey.PublicKey), Data: dataMessage}
	from.WriteRequest(from.Friends[0], request)
}
