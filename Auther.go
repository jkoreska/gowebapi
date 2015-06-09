package gowebapi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/jameskeane/bcrypt"
	"strings"
	"time"
)

type Auther interface {
	Authenticate(request *Request, response *Response) (*Response)
	Signin(userdata string, expiryMinutes int64) string
	Hash(password string) string
	Match(password string, hash string) bool
}

type defaultAuther struct {
	key []byte
}

func NewDefaultAuther(key []byte) Auther {
	return &defaultAuther{key}
}

func (self *defaultAuther) Authenticate(request *Request, response *Response) (*Response) {

	if nil != response {
		return nil
	}

	authHeaders, authExists := request.Http.Header["Authorization"]

	if authExists {

		header := strings.Trim(authHeaders[0], " ")
		parts := strings.Split(header, " ")

		if 2 == len(parts) {

			auth, decodeErr := base64.StdEncoding.DecodeString(parts[1])

			if nil == decodeErr && "basic" == strings.ToLower(parts[0][:5]) {

				authParts := strings.Split(string(auth), ":")

				userData := self.decodeTicket(authParts[0])

				if "" != userData {
					request.UserData = userData

					return nil
				}
			}
		}
	}

	return &Response{
		Status: 401,
		Header: map[string][]string{"Www-Authenticate": []string{"Basic/WebAPI"}},
	}
}

func (self *defaultAuther) Signin(userdata string, expiryMinutes int64) string {

	return self.encodeTicket(userdata, expiryMinutes)
}

func (self *defaultAuther) Hash(password string) string {

	hash, _ := bcrypt.Hash(password)

	return hash
}

func (self *defaultAuther) Match(password string, hash string) bool {

	return bcrypt.Match(password, hash)
}

func (self *defaultAuther) encodeTicket(userdata string, expiryMinutes int64) string {

	// generate the token uuid|userdata|expiry

	uuid := Uuid()
	time := time.Now().Add(time.Duration(expiryMinutes) * time.Minute).Format(time.RFC3339)

	token := fmt.Sprintf("%s|%s|%s", uuid, userdata, time)

	// encrypt the token

	block, blockError := aes.NewCipher(self.key)

	if nil != blockError {
		return ""
	}

	bytes := []byte(token)
	encrypted := make([]byte, len(bytes))

	encrypter := cipher.NewCFBEncrypter(block, self.key[:aes.BlockSize])
	encrypter.XORKeyStream(encrypted, bytes)

	return hex.EncodeToString(encrypted)
}

func (self *defaultAuther) decodeTicket(token string) string {

	if "" != token {

		// decrypt the token

		bytes, decodeError := hex.DecodeString(token)

		if nil != decodeError {
			return ""
		}

		block, blockError := aes.NewCipher(self.key)

		if nil != blockError {
			return ""
		}

		decrypted := make([]byte, len(bytes))

		decrypter := cipher.NewCFBDecrypter(block, self.key[:aes.BlockSize])
		decrypter.XORKeyStream(decrypted, bytes)

		// split the decrypted string into uuid|userdata|expiry

		parts := strings.Split(string(decrypted), "|")

		if 3 != len(parts) {
			return ""
		}

		// TODO: validate the expiry
		// TODO: handle 0 (infinite) expiry

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

func Uuid() string {
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
