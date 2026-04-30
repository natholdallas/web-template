// Package mail to send email
package mail

import (
	"net/smtp"

	"webtplmst/internal/conf"
)

func SendMail(to []string, msg []byte) error {
	auth := smtp.PlainAuth("", conf.App.SMTPFrom, conf.App.SMTPPassword, conf.App.SMTPHost)
	return smtp.SendMail(conf.App.SMTPAddr, auth, conf.App.SMTPFrom, to, msg)
}
