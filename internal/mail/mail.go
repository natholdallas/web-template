// Package mail to send email
package mail

// SendMail dials the configured SMTP server and delivers the message in one
// shot, closing the connection afterwards.
func SendMail(to []string, subject string, build func(*Message)) error {
	m := NewMessage(to, subject)
	build(m)
	cl, err := NewClient()
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Send(m)
}
