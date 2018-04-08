package sh

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
)

func createToken(key string, token Token) string {
	//First serialize the token into json string
	b, err := json.Marshal(&token)
	if err != nil {
		log.Println(err)
		return ""
	}
	tosign := "eyJhbGciOiJIUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(b)
	//Generate an hmac on that string
	sig := hmac.New(sha256.New, []byte(key))
	_, err = sig.Write([]byte(tosign))
	if err != nil {
		log.Println(err)
		return ""
	}
	//Base64 the string and the hmac
	//Return them as [jwtheader].[Base64encodedmessage].[base64encodedhmac]
	return tosign + "." + base64.RawURLEncoding.EncodeToString(sig.Sum(nil))
}

func validateToken(key, token string) (Token, error) {
	ret := Token{}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ret, errors.New("Invalid Token")
	}
	if parts[0] != "eyJhbGciOiJIUzI1NiJ9" {
		return ret, errors.New("Invalid Header")
	}
	//First decode the message, calc signature
	sig := hmac.New(sha256.New, []byte(key))
	_, err := sig.Write([]byte(parts[0] + "." + parts[1]))
	if err != nil {
		log.Println(err)
		return ret, errors.New("Invalid Signature")
	}
	if parts[2] != base64.RawURLEncoding.EncodeToString(sig.Sum(nil)) {
		return ret, errors.New("Invalid Signature")
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
