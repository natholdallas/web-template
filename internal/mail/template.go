package mail

import (
	"fmt"
	"math/rand"

	"webtplmst/internal/conf"
)

func VerficationCodeTPL(name, email, code string) []byte {
	subject := "Verification Code"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>%s Verification Code</title>
</head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #1a1a1a; color: #f1f1f1; margin:0; padding:0;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background: linear-gradient(135deg, #6a11cb 0%%, #2575fc 100%%); min-height: 100vh; padding: 40px;">
    <tr>
      <td align="center" style="padding: 50px 0;">
        <table width="600" cellpadding="0" cellspacing="0" style="background: #fff; color: #333; border-radius: 10px; padding: 40px; box-shadow: 0 0 20px rgba(0,0,0,0.3);">
  	 			<tr>
      			<td align="center">
        			<h1 style="font-size: 28px; color: #2575fc;">%s Verification Code</h1>
        			<p style="font-size: 16px;">Hello, your verification code is as follows. Please use it within <strong>5 minutes</strong></p>
        			<div style="margin: 30px 0; font-size: 32px; font-weight: bold; color: #ff5722; letter-spacing: 4px; border: 2px dashed #ff5722; padding: 15px 0; width: 250px; text-align: center; border-radius: 8px;">%s</div>
        			<p style="margin-top: 30px; font-size: 12px; color: #999;">If it is not your operation, please ignore this email</p>
      			</td>
    			</tr>
  			</table>
      </td>
    </tr>
  </table>
</body>`, name, name, code)

	// 邮件头
	return fmt.Appendf(nil,
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n%s",
		conf.App.SMTPFrom, email, subject, body,
	)
}

func GenerateVerificationCode() string {
	min := 100000
	max := 999999
	rangeSize := max - min + 1
	code := rand.Intn(rangeSize) + min
	return fmt.Sprint(code)
}
