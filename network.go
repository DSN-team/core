package main

type NetworkInterface interface {
	startTimer()
	sendData(callback func())
}
