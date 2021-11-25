package main

import (
	"github.com/DSN-team/core"
	"github.com/DSN-team/core/tests/utils"
	"strconv"
	"time"
)

func connectNodes(profiles []*core.Profile, i, j int) {
	utils.CreateNetwork(profiles[i], profiles[j])
	utils.CreateNetwork(profiles[j], profiles[i])
}
func main() {
	utils.InitTest()
	profiles := make([]*core.Profile, 8)
	for i := 0; i < len(profiles); i++ {
		profiles[i] = utils.RunProfile(strconv.FormatInt(int64(i), 10))
	}
	connectNodes(profiles, 0, 1)
	connectNodes(profiles, 0, 2)
	connectNodes(profiles, 0, 3)
	connectNodes(profiles, 1, 6)
	connectNodes(profiles, 6, 2)
	connectNodes(profiles, 3, 4)
	connectNodes(profiles, 5, 4)
	connectNodes(profiles, 6, 7)
	//for i := 0; i < len(profiles); i++ {
	//	go utils.StartConnection(profiles[i])
	//}

	//Hold main thread
	for {
		time.Sleep(10)
	}
}
