package core

import "testing"

func TestAES(t *testing.T) {
	profileKey := genProfileKey()
	SelectedProfile.PrivateKey = profileKey
	publicKey := profileKey.PublicKey
	println("PublicKey:", (publicKey).X.Uint64())
	testStr := "12312"
	if string(decryptAES(encryptAES(&publicKey, []byte(testStr)))) != testStr {
		t.Error("AES failed")
	}
}
