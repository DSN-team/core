package core

import "testing"

func TestAES(t *testing.T) {
	SelectedProfile := Profile{}
	profileKey := genProfileKey()
	SelectedProfile.PrivateKey = profileKey
	publicKey := profileKey.PublicKey
	println("PublicKey:", (publicKey).X.Uint64())
	testStr := "12312"
	if string(SelectedProfile.decryptAES(SelectedProfile.decryptAES([]byte(testStr)))) != testStr {
		t.Error("AES failed")
	}
}
