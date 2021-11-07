package core

import "testing"

func TestRegister(t *testing.T) {
	TestStartDB(t)
	SelectedProfile := Profile{}
	result := SelectedProfile.Register("test", "test", "127.0.0.1:3000")
	if result == false {
		t.Error("Register failed")
	}
}

func TestLogin(t *testing.T) {
	TestStartDB(t)
	SelectedProfile := Profile{}
	LoadProfiles()
	result := SelectedProfile.Login("test", 1)
	if result == false {
		t.Error("Login failed")
	}
}
