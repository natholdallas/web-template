package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"

	"webtplmst/internal/conf"
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

func VerficationCodeTPL(name, email, code string) []byte {
	subject := "Verification Code"
	// Execute the HTML template into a buffer
	var body bytes.Buffer
	data := map[string]string{
		"Name": name,
		"Code": code,
	}
	// Error is ignored here as the template is pre-validated
	_ = vCodeTmpl.Execute(&body, data)
	// Encode the Subject using RFC 2047 (Base64) to prevent encoding issues in mail clients
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	// Construct the raw email byte stream using a Buffer to minimize memory allocations
	res := bytes.NewBuffer(nil)
	fmt.Fprintf(res, "From: %s\r\n", conf.App.SMTPFrom)
	fmt.Fprintf(res, "To: %s\r\n", email)
	fmt.Fprintf(res, "Subject: %s\r\n", encodedSubject)
	res.WriteString("MIME-Version: 1.0\r\n")
	res.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	res.WriteString("\r\n")
	res.Write(body.Bytes())
	return res.Bytes()
}

func GenerateVerificationCode() string {
	min := 100000
	max := 999999
	rangeSize := max - min + 1
	code := rand.Intn(rangeSize) + min
	return fmt.Sprint(code)
}
