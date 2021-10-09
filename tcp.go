package main

type TcpStruct struct {
}

func (u TcpStruct) sendData(callback func()) {
	panic("implement me")
}

func (u TcpStruct) startTimer() {
	println("There's no timer fo TCP")
}
