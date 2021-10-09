package main

import "time"

type UdpStruct struct {
}

func (u UdpStruct) sendData(callback func()) {
	panic("implement me")
}

func (u UdpStruct) startTimer() {
	go func() {
		for {
			doUdpStuff()
			time.Sleep(150)
		}
	}()
}

func doUdpStuff() {
	println("doing Udp stuff")
}
