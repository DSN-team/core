package utils

import (
	"fmt"
	"github.com/DSN-team/core"
	utils2 "github.com/DSN-team/core/utils"
	"time"
)

func DelayedCall(from, to *core.Profile, msg string) {
	time.Sleep(200 * time.Millisecond)
	core.UpdateUI = func(i int, client int) {
		fmt.Print("client:", client, "\n")
		fmt.Print("got it:", to.DataStrOutput)
		fmt.Println("got it as string:", string(to.DataStrOutput))
	}

	for i := 0; i < len(msg); i++ {
		from.DataStrInput[i] = msg[i]
	}
	dataMessage := from.BuildDataMessage([]byte(msg), from.Friends[0].ID)
	request := core.Request{RequestType: utils2.RequestData, PublicKey: core.MarshalPublicKey(&from.PrivateKey.PublicKey), Data: dataMessage}
	from.WriteRequest(from.Friends[0], request)
}
