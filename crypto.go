package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"crypto/x509"
	"fmt"
	"log"
)

var Curve = elliptic.P256

var aesBlockLength = aes.BlockSize

func genProfileKey() *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(Curve(), rand.Reader)
	ErrHandler(err)
	fmt.Println(key.PublicKey)
	return key
}

func encProfileKey() (data []byte) {
	passwordHash := sha512.Sum512_256([]byte(profile.password))
	iv := passwordHash[:aesBlockLength]
	log.Println(len(iv), ": ", iv)
	log.Println(len(passwordHash), ": ", passwordHash)
	ErrHandler(err)
	key, err := x509.MarshalECPrivateKey(profile.privateKey)
	ErrHandler(err)
	data, err = encryptCBC(key, iv, passwordHash[:aesBlockLength])
	ErrHandler(err)
	return
}

func decProfileKey(encKey []byte, password string) bool {
	passwordHash := sha512.Sum512_256([]byte(password))
	iv := passwordHash[:aesBlockLength]
	data := decryptCBC(encKey, iv, passwordHash[:aesBlockLength])
	profile.privateKey, err = x509.ParseECPrivateKey(data)

	if ErrHandler(err) {
		return false
	} else {
		return true
	}
}

func encryptAES(otherPublicKey *ecdsa.PublicKey, in []byte) (out []byte) {
	x, _ := otherPublicKey.Curve.ScalarMult(otherPublicKey.X, otherPublicKey.Y, profile.privateKey.D.Bytes())
	if x == nil {
		return nil
	}
	shared := sha512.Sum512_256(x.Bytes())
	iv, err := makeRandom(aesBlockLength)
	if ErrHandler(err) {
		return
	}

	ct, err := encryptCBC(in, iv, shared[:])
	if ErrHandler(err) {
		return
	}

	ephPub := elliptic.Marshal(otherPublicKey.Curve, profile.privateKey.PublicKey.X, profile.privateKey.PublicKey.Y)
	out = make([]byte, 1+len(ephPub)+aesBlockLength)
	out[0] = byte(len(ephPub))
	copy(out[1:], ephPub)
	copy(out[1+len(ephPub):], iv)
	out = append(out, ct...)

	h := hmac.New(sha1.New, shared[:])
	h.Write(iv)
	h.Write(ct)
	out = h.Sum(out)
	return
}

func decryptAES(in []byte) (out []byte) {
	ephLen := int(in[0])
	ephPub := in[1 : 1+ephLen]
	ct := in[1+ephLen:]
	if len(ct) < (sha1.Size + aesBlockLength) {
		return nil
	}

	x, y := elliptic.Unmarshal(Curve(), ephPub)
	ok := Curve().IsOnCurve(x, y) // Rejects the identity point too.
	if x == nil || !ok {
		return nil
	}

	x, _ = profile.privateKey.Curve.ScalarMult(x, y, profile.privateKey.D.Bytes())
	if x == nil {
		return nil
	}
	shared := sha512.Sum512_256(x.Bytes())

	tagStart := len(ct) - sha1.Size
	h := hmac.New(sha1.New, shared[:])
	h.Write(ct[:tagStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, ct[tagStart:]) {
		return nil
	}

	out = decryptCBC(ct[aes.BlockSize:tagStart], ct[:aes.BlockSize], shared[:])
	return
}

func encryptCBC(data, iv, key []byte) (encryptedData []byte, err error) {
	paddedData := addPadding(data)
	fmt.Println("paddedata: ", paddedData)
	fmt.Println("data len:", len(paddedData))

	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return nil, err
	}

	fmt.Println("block size:", aesCrypt.BlockSize())

	encryptedData = make([]byte, len(paddedData))
	aesCBC := cipher.NewCBCEncrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(encryptedData, paddedData)

	return
}

func decryptCBC(data, iv, key []byte) (decryptedData []byte) {
	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return
	}

	decryptedPaddedData := make([]byte, len(data))
	aesCBC := cipher.NewCBCDecrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(decryptedPaddedData, data)
	decryptedData = removePadding(decryptedPaddedData)
	fmt.Println("decrypted depadded data:", decryptedData)
	return
}

func makeRandom(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

func removePadding(data []byte) []byte {
	fmt.Println("padded data len:", len(data))
	if len(data) == 0 {
		return nil
	}
	l := int(data[len(data)-1])
	if l > aesBlockLength {
		return nil
	}
	return data[:len(data)-l]
}

// addPadding adds padding to a block of data
func addPadding(b []byte) []byte {
	l := aesBlockLength - len(b)%aesBlockLength
	padding := make([]byte, l)
	padding[l-1] = byte(l)
	return append(b, padding...)
}
