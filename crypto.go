package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"errors"
	"fmt"
)

var Curve = elliptic.P256

func genProfileKey() *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(Curve(), rand.Reader)
	ErrHandler(err)
	fmt.Println(key.PublicKey)
	return key
}

func encProfileKey() (data []byte) {
	passwordHash := sha512.Sum512([]byte(profile.password))
	iv := passwordHash[16:]
	ErrHandler(err)
	key, err := x509.MarshalECPrivateKey(profile.privateKey)
	ErrHandler(err)
	data, err = encryptCBC(key, iv, passwordHash[:16])
	ErrHandler(err)
	return
}

func decProfileKey(encKey []byte, password string) {
	passwordHash := sha512.Sum512([]byte(profile.password))
	iv := passwordHash[:16]
	data, err := decryptCBC(encKey, passwordHash[:16], iv)
	ErrHandler(err)
	profile.privateKey, err = x509.ParseECPrivateKey(data)
}

func encryptAES(otherPublicKey *ecdsa.PublicKey, in []byte) (out []byte, err error) {
	x, _ := otherPublicKey.Curve.ScalarMult(otherPublicKey.X, otherPublicKey.Y, profile.privateKey.D.Bytes())
	if x == nil {
		return nil, errors.New("Failed to generate encryption privateKey")
	}
	shared := sha256.Sum256(x.Bytes())
	iv, err := makeRandom(16)
	if err != nil {
		return
	}

	paddedIn := addPadding(in)
	ct, err := encryptCBC(paddedIn, iv, shared[:16])
	if err != nil {
		return
	}

	ephPub := elliptic.Marshal(otherPublicKey.Curve, profile.privateKey.PublicKey.X, profile.privateKey.PublicKey.Y)
	out = make([]byte, 1+len(ephPub)+16)
	out[0] = byte(len(ephPub))
	copy(out[1:], ephPub)
	copy(out[1+len(ephPub):], iv)
	out = append(out, ct...)

	h := hmac.New(sha1.New, shared[16:])
	h.Write(iv)
	h.Write(ct)
	out = h.Sum(out)
	return
}

func decryptAES(in []byte) (out []byte, err error) {
	ephLen := int(in[0])
	ephPub := in[1 : 1+ephLen]
	ct := in[1+ephLen:]
	if len(ct) < (sha1.Size + aes.BlockSize) {
		return nil, errors.New("Invalid ciphertext")
	}

	x, y := elliptic.Unmarshal(Curve(), ephPub)
	ok := Curve().IsOnCurve(x, y) // Rejects the identity point too.
	if x == nil || !ok {
		return nil, errors.New("Invalid public privateKey")
	}

	x, _ = profile.privateKey.Curve.ScalarMult(x, y, profile.privateKey.D.Bytes())
	if x == nil {
		return nil, errors.New("Failed to generate encryption privateKey")
	}
	shared := sha256.Sum256(x.Bytes())

	tagStart := len(ct) - sha1.Size
	h := hmac.New(sha1.New, shared[16:])
	h.Write(ct[:tagStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, ct[tagStart:]) {
		return nil, errors.New("Invalid MAC")
	}

	paddedOut, err := decryptCBC(ct[aes.BlockSize:tagStart], ct[:aes.BlockSize], shared[:16])
	if err != nil {
		return
	}
	out, err = removePadding(paddedOut)
	return
}

func encryptCBC(data, iv, key []byte) (encryptedData []byte, err error) {
	aesCrypt, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ivBytes := append([]byte{}, iv...)

	encryptedData = make([]byte, len(data))
	aesCBC := cipher.NewCBCEncrypter(aesCrypt, ivBytes)
	aesCBC.CryptBlocks(encryptedData, data)

	return
}

func decryptCBC(data, iv, key []byte) (decryptedData []byte, err error) {
	aesCrypt, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ivBytes := append([]byte{}, iv...)

	decryptedData = make([]byte, len(data))
	aesCBC := cipher.NewCBCDecrypter(aesCrypt, ivBytes)
	aesCBC.CryptBlocks(decryptedData, data)

	return
}

func makeRandom(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

func removePadding(b []byte) ([]byte, error) {
	l := int(b[len(b)-1])
	if l > 16 {
		return nil, errors.New("Padding incorrect")
	}

	return b[:len(b)-l], nil
}

// addPadding adds padding to a block of data
func addPadding(b []byte) []byte {
	l := 16 - len(b)%16
	padding := make([]byte, l)
	padding[l-1] = byte(l)
	return append(b, padding...)
}
