package mail

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"html/template"
	"math/big"
)

var vCodeTmpl = template.Must(template.New("vcode").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: sans-serif; background: #1a1a1a; color: #f1f1f1; margin:0; padding:0;">
  <table width="100%" style="background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%); padding: 40px;">
    <tr>
      <td align="center">
        <table width="600" style="background: #fff; color: #333; border-radius: 10px; padding: 40px;">
          <tr>
            <td align="center">
              <h1 style="color: #2575fc;">{{.Name}} Verification Code</h1>
              <p>Hello, your verification code is as follows. Please use it within <strong>5 minutes</strong></p>
              <div style="margin: 30px 0; font-size: 32px; font-weight: bold; color: #ff5722; letter-spacing: 4px; border: 2px dashed #ff5722; padding: 15px 0; width: 250px; border-radius: 8px;">{{.Code}}</div>
              <p style="font-size: 12px; color: #999;">If it is not your operation, please ignore this email</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`))

// VerificationCodeTPL builds a verification-code message for the given name,
// email and code.
func VerificationCodeTPL(name, email, code string) *Message {
	var body bytes.Buffer
	vCodeTmpl.Execute(&body, map[string]string{"Name": name, "Code": code})
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
