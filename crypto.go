package core

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
	"encoding/base64"
	"log"
	"math/big"
)

var Curve = elliptic.P256

var aesBlockLength = aes.BlockSize

func genProfileKey() (key *ecdsa.PrivateKey) {
	key, err = ecdsa.GenerateKey(Curve(), rand.Reader)
	ErrHandler(err)
	return
}

func EncPublicKey(key []byte) (keyString string) {
	log.Println("encoding public key")
	keyString = base64.StdEncoding.EncodeToString(key)
	log.Println("encoded public key", keyString)
	return
}

func DecPublicKey(data string) (publicKeyBytes []byte) {
	println("decoding public key")
	publicKeyBytes, err = base64.StdEncoding.DecodeString(data)
	if ErrHandler(err) {
		return nil
	}
	log.Println("decode public key", publicKeyBytes)
	return
}

func MarshalPublicKey(key *ecdsa.PublicKey) (data []byte) {
	println("marshal Public Key", key)
	data = elliptic.Marshal(key, key.X, key.Y)
	log.Println("marshalled Public Key", data)
	return
}

func UnmarshalPublicKey(data []byte) (key ecdsa.PublicKey) {
	println("unmarshal Public Key", data)
	x, y := elliptic.Unmarshal(Curve(), data)
	key = ecdsa.PublicKey{Curve: Curve(), X: x, Y: y}
	log.Println("unmarshalled public key", key)
	return
}

func (profile *Profile) encProfileKey() (data []byte) {
	log.Println("Encrypting Profile key")
	passwordHash := sha512.Sum512_256([]byte(profile.Password))
	iv := passwordHash[:aes.BlockSize]
	key, err := x509.MarshalECPrivateKey(profile.PrivateKey)
	ErrHandler(err)
	data = encryptCBC(key, iv, passwordHash[:aes.BlockSize])
	return
}

func (profile *Profile) decProfileKey(encKey []byte, password string) bool {
	log.Println("Decrypting Profile key")
	passwordHash := sha512.Sum512_256([]byte(password))
	iv := passwordHash[:aes.BlockSize]
	data := decryptCBC(encKey, iv, passwordHash[:aes.BlockSize])
	profile.PrivateKey, err = x509.ParseECPrivateKey(data)

	if ErrHandler(err) {
		return false
	} else {
		return true
	}
}

func (profile *Profile) encryptAES(otherPublicKey *ecdsa.PublicKey, in []byte) (out []byte) {
	log.Println("Encrypting aes")
	x, _ := otherPublicKey.Curve.ScalarMult(otherPublicKey.X, otherPublicKey.Y, profile.PrivateKey.D.Bytes())
	if x == nil {
		return nil
	}
	shared := sha512.Sum512_256(x.Bytes())
	iv, err := makeRandom(aesBlockLength)
	if ErrHandler(err) {
		return
	}

	ct := encryptCBC(in, iv, shared[:])
	if ct == nil {
		return
	}

	ephPub := elliptic.Marshal(otherPublicKey.Curve, profile.PrivateKey.PublicKey.X, profile.PrivateKey.PublicKey.Y)
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

func (profile *Profile) decryptAES(in []byte) (out []byte) {
	log.Println("Decrypting aes")
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

	x, _ = profile.PrivateKey.Curve.ScalarMult(x, y, profile.PrivateKey.D.Bytes())
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

func encryptCBC(data, iv, key []byte) []byte {
	log.Println("Encrypting cbc")
	paddedData := addPadding(data)
	log.Println("padded data: ", paddedData)
	log.Println("data len:", len(paddedData))

	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return nil
	}

	log.Println("block size:", aesCrypt.BlockSize())

	encryptedData := make([]byte, len(paddedData))
	aesCBC := cipher.NewCBCEncrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(encryptedData, paddedData)

	return encryptedData
}

func decryptCBC(data, iv, key []byte) (decryptedData []byte) {
	log.Println("Decrypting cbc")
	log.Println("data length:", len(data), "data:", "iv:", iv, "key:", key)
	aesCrypt, err := aes.NewCipher(key)
	if ErrHandler(err) {
		return
	}

	decryptedPaddedData := make([]byte, len(data))
	aesCBC := cipher.NewCBCDecrypter(aesCrypt, iv)
	aesCBC.CryptBlocks(decryptedPaddedData, data)
	log.Println("decrypted padded data len:", len(decryptedPaddedData), "decrypted padded data:", decryptedPaddedData)
	decryptedData = removePadding(decryptedPaddedData)
	log.Println("decrypted de padded data:", decryptedData)
	return
}

func makeRandom(length int) ([]byte, error) {
	log.Println("Making random:", length)
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

func removePadding(data []byte) []byte {
	log.Println("Removing padding")
	log.Println("padded data len:", len(data))
	if len(data) == 0 {
		return nil
	}
	l := int(data[len(data)-1])
	log.Println("len of data:", l)
	log.Println("real len of data:", len(data))
	if l > len(data) {
		return nil
	} else {
		return data[:len(data)-l]
	}
}

// addPadding adds padding to a block of data
func addPadding(data []byte) []byte {
	log.Println("Adding padding")
	log.Println("data len:", len(data))
	l := aesBlockLength - len(data)%aesBlockLength
	log.Println("additional len to pad data", l)
	padding := make([]byte, l)
	padding[l-1] = byte(l)
	data = append(data, padding...)
	return data
}

func (profile *Profile) signData(data []byte) (*big.Int, *big.Int) {
	log.Println("Signing data")
	r, s, err := ecdsa.Sign(rand.Reader, profile.PrivateKey, data)
	ErrHandler(err)
	return r, s
}

func (profile *Profile) verifyData(data []byte, r, s big.Int) (result bool) {
	log.Println("Verify data")
	return ecdsa.Verify(&profile.PrivateKey.PublicKey, data, &r, &s)
}
