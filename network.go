package dsn_go

type NetworkInterface interface {
	startTimer()
	sendData(callback func())
}
