package core

type NetworkInterface interface {
	startTimer()
	sendData(callback func())
}
