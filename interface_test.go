package core

import "testing"

func TestRegister(t *testing.T) {
	TestStartDB(t)
	result := Register("test", "test")
	if result == false {
		t.Error("Register failed")
	}
}

func TestLogin(t *testing.T) {
	TestStartDB(t)
	result := Login("test", 1)
	if result == false {
		t.Error("Login failed")
	}
}
