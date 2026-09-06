package mail

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
)

// VerificationCodeTPL builds a verification-code message for the given name,
// email and code.
func VerificationCodeTPL(name, email, code string) *Message {
	var body bytes.Buffer
	verficationCodeTpl.Execute(&body, map[string]string{"Name": name, "Code": code})
	m := NewMessage([]string{email}, "Verification Code")
	m.SetBodyHTML(body.String())
	return m
}

// GenerateVerificationCode returns a cryptographically random 6-digit code.
func GenerateVerificationCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}
