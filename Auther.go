package gowebapi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/jameskeane/bcrypt"
	"strings"
	"time"
)

var key = []byte{0x6c, 0xf8, 0x05, 0x1b, 0x4a, 0xae, 0xc0, 0xa9, 0x7f, 0x47, 0x94, 0x8d, 0x11, 0xdf, 0xe0, 0x0a}

type Auther interface {
	Authorize(request *Request) bool
	GenerateToken(userdata string, expiryMinutes int64) string
	AuthenticateToken(token string) string
}

type DefaultAuther struct{}

func (self *DefaultAuther) Authorize(request *Request) bool {

	if "" == request.Authorize {
		return true
	}

	tokens, tokenExists := request.Http.Header["Authorize"]

	if !tokenExists {
		return false
	}

	request.UserData = self.AuthenticateToken(tokens[0])

	return "" != request.UserData
}

func (self *DefaultAuther) GenerateToken(userdata string, expiryMinutes int64) string {

	// generate the token uuid|userdata|expiry

	uuid := self.makeUUID()
	time := time.Now().Add(time.Duration(expiryMinutes) * time.Minute).Format(time.RFC3339)

	token := fmt.Sprintf("%s|%s|%s", uuid, userdata, time)

	// encrypt the token

	block, blockError := aes.NewCipher(key)

	if nil != blockError {
		return ""
	}

	bytes := []byte(token)
	encrypted := make([]byte, len(bytes))

	encrypter := cipher.NewCFBEncrypter(block, key[:aes.BlockSize])
	encrypter.XORKeyStream(encrypted, bytes)

	return hex.EncodeToString(encrypted)
}

func (self *DefaultAuther) AuthenticateToken(token string) string {

	if "" != token {

		// decrypt the token

		bytes, decodeError := hex.DecodeString(token)

		if nil != decodeError {
			return ""
		}

		block, blockError := aes.NewCipher(key)

		if nil != blockError {
			return ""
		}

		decrypted := make([]byte, len(bytes))

		decrypter := cipher.NewCFBDecrypter(block, key[:aes.BlockSize])
		decrypter.XORKeyStream(decrypted, bytes)

		// split the decrypted string into uuid|userdata|expiry

		parts := strings.Split(string(decrypted), "|")

		if 3 != len(parts) {
			return ""
		}

		// validate the expiry

		expiry, expiryError := time.Parse(time.RFC3339, parts[2])

		if nil != expiryError {
			return ""
		}

		if time.Now().Sub(expiry) > 0 {
			return ""
		}

		// return userdata
		return parts[1]

	} else {

		return ""
	}
}

func (self *DefaultAuther) GenerateSalt() string {

	salt, _ := bcrypt.Salt()

	return salt
}

func (self *DefaultAuther) HashPassword(password string, salt string) string {

	hash, _ := bcrypt.Hash(password, salt)

	return hash
}

func (self *DefaultAuther) makeUUID() string {
	// http://stackoverflow.com/questions/15130321

	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)

	if err != nil {
		return ""
	}

	bytes[8] = (bytes[8] | 0x80) & 0xBF // identify UUID V4
	bytes[6] = (bytes[6] | 0x40) & 0x4F //

	return hex.EncodeToString(bytes)
}
