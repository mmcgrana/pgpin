package main

import (
	"github.com/fernet/fernet-go"
)

func FernetEncrypt(s string) []byte {
	tok, err := fernet.EncryptAndSign([]byte(s), ConfigFernetKeys[0])
	Must(err)
	return tok
}

func FernetDecrypt(b []byte) string {
	msg := fernet.VerifyAndDecrypt(b, ConfigFernetTtl, ConfigFernetKeys)
	return string(msg)
}
