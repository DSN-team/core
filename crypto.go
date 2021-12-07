package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
)

var Curve = elliptic.P256

var aesBlockLength = aes.BlockSize

type DataAES struct {
	Data []byte
	IV   []byte
}

func genProfileKey() (key *ecdsa.PrivateKey) {
	fmt.Println("Generating profile key...")
	key, err = ecdsa.GenerateKey(Curve(), rand.Reader)
	ErrHandler(err)
	return
}

func EncodeKey(key []byte) (keyString string) {
	fmt.Println("Encoding key")
	keyString = base64.StdEncoding.EncodeToString(key)
	return
}

func DecodeKey(data string) (keyBytes []byte) {
	fmt.Println("Decoding key")
	keyBytes, err = base64.StdEncoding.DecodeString(data)
	if ErrHandler(err) {
		return nil
	}
	return
}

func MarshalPublicKey(key *ecdsa.PublicKey) (data []byte) {
	fmt.Println("Marshal Public Key", key)
	data = elliptic.Marshal(key, key.X, key.Y)
	return
}

func UnmarshalPublicKey(data []byte) (key ecdsa.PublicKey) {
	fmt.Println("Unmarshal Public Key", data)
	x, y := elliptic.Unmarshal(Curve(), data)
	key = ecdsa.PublicKey{Curve: Curve(), X: x, Y: y}
	return
}

func (cur *Profile) encProfileKey() {
	fmt.Println("Encrypting profile key")
	passwordHash := sha512.Sum512_256([]byte(cur.Password))
	iv := passwordHash[:aes.BlockSize]
	key, err := x509.MarshalECPrivateKey(cur.PrivateKey)
	ErrHandler(err)
	cur.PrivateKeyString = EncodeKey(encryptCBC(key, iv, passwordHash[:aes.BlockSize]))
	return
}

func (cur *Profile) decProfileKey(encKey string, password string) bool {
	fmt.Println("Decrypting profile key")
	passwordHash := sha512.Sum512_256([]byte(password))
	iv := passwordHash[:aes.BlockSize]
	data := decryptCBC(DecodeKey(encKey), iv, passwordHash[:aes.BlockSize])
	cur.PrivateKey, err = x509.ParseECPrivateKey(data)

	if ErrHandler(err) {
		return false
	} else {
		return true
	}
}

func (cur *Profile) encryptAES(otherPublicKey *ecdsa.PublicKey, in []byte) (out []byte) {
	fmt.Println("Encrypting aes")

	scalarX, _ := otherPublicKey.Curve.ScalarMult(otherPublicKey.X, otherPublicKey.Y, cur.PrivateKey.D.Bytes())
	if scalarX == nil {
		return nil
	}
	sharedKey := sha512.Sum512_256(scalarX.Bytes())

	iv, err := makeRandom(aesBlockLength)
	if ErrHandler(err) {
		return
	}

	encryptedData := encryptCBC(in, iv, sharedKey[:])
	if encryptedData == nil {
		return
	}

	var dataAES DataAES
	var dataAESBuffer bytes.Buffer

	dataAES.IV = iv
	dataAES.Data = encryptedData

	dataAESEncoder := gob.NewEncoder(&dataAESBuffer)
	err = dataAESEncoder.Encode(dataAES)
	if ErrHandler(err) {
		return
	}

	dataAESBytes := dataAESBuffer.Bytes()

	h := hmac.New(sha512.New, sharedKey[:])
	h.Write(dataAESBytes)
	out = h.Sum(dataAESBytes)
	return
}

func (cur *Profile) decryptAES(otherPublicKey *ecdsa.PublicKey, in []byte) (out []byte) {
	fmt.Println("Decrypting aes")

	scalarX, _ := cur.PrivateKey.Curve.ScalarMult(otherPublicKey.X, otherPublicKey.Y, cur.PrivateKey.D.Bytes())
	if scalarX == nil {
		return nil
	}
	shared := sha512.Sum512_256(scalarX.Bytes())

	hashStart := len(in) - sha512.Size
	h := hmac.New(sha512.New, shared[:])
	h.Write(in[:hashStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, in[hashStart:]) {
		log.Println("AES checksum mismatch")
		return nil
	}

	var dataAES DataAES
	dataAESBuffer := bytes.NewBuffer(in)

	dataAESDecoder := gob.NewDecoder(dataAESBuffer)
	err := dataAESDecoder.Decode(&dataAES)
	if ErrHandler(err) {
		return
	}

	out = decryptCBC(dataAES.Data, dataAES.IV, shared[:])
	return
}

func encryptCBC(data, iv, key []byte) []byte {
	fmt.Println("Encrypting cbc")
	paddedData := addPadding(data)
	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return nil
	}

	encryptedData := make([]byte, len(paddedData))
	aesCBC := cipher.NewCBCEncrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(encryptedData, paddedData)

	return encryptedData
}

func decryptCBC(data, iv, key []byte) (decryptedData []byte) {
	fmt.Println("Decrypting cbc")
	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return
	}

	decryptedPaddedData := make([]byte, len(data))
	aesCBC := cipher.NewCBCDecrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(decryptedPaddedData, data)
	decryptedData = removePadding(decryptedPaddedData)
	return
}

func makeRandom(length int) ([]byte, error) {
	fmt.Println("Making random:", length)
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	return buffer, err
}

func removePadding(data []byte) []byte {
	fmt.Println("Removing padding")
	if len(data) == 0 {
		return nil
	}
	l := int(data[len(data)-1])
	if l > len(data) {
		return nil
	} else {
		return data[:len(data)-l]
	}
}

// addPadding adds padding to a block of data
func addPadding(data []byte) []byte {
	fmt.Println("Adding padding")
	l := aesBlockLength - len(data)%aesBlockLength
	padding := make([]byte, l)
	padding[l-1] = byte(l)
	data = append(data, padding...)
	return data
}

func (cur *Profile) signData(data []byte) (r *big.Int, s *big.Int) {
	fmt.Print("Signing data")
	r, s, err = ecdsa.Sign(rand.Reader, cur.PrivateKey, data)
	ErrHandler(err)
	fmt.Println("data signed, r:", r, "s:", s)
	return
}

func (cur *Profile) verifyData(data []byte, r, s big.Int) (result bool) {
	fmt.Print("Verify data, r:", r, "s:", s)
	result = ecdsa.Verify(&cur.PrivateKey.PublicKey, data, &r, &s)
	fmt.Println("data valid:", result)
	return
}
