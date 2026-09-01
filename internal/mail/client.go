package mail

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"webtplmst/internal/conf"
)

// Client is a reusable SMTP client. It holds a single persistent connection so
// multiple messages can be sent without re-establishing the session each time.
type Client struct {
	c    *smtp.Client
	from string
}

// NewClient dials the configured SMTP server, upgrades to TLS where required
// and authenticates. Port 465 uses implicit TLS; other ports use STARTTLS when
// the server advertises it, otherwise plaintext.
func NewClient() (*Client, error) {
	host, _, _ := net.SplitHostPort(conf.App.SMTPAddr)
	var c *smtp.Client

	if conf.App.SMTPPort == 465 {
		conn, err := tls.Dial("tcp", conf.App.SMTPAddr, &tls.Config{ServerName: host})
		if err != nil {
			return nil, err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			conn.Close()
			return nil, err
		}
	} else {
		var err error
		c, err = smtp.Dial(conf.App.SMTPAddr)
		if err != nil {
			return nil, err
		}
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
				c.Close()
				return nil, err
			}
		}
	}

	auth := smtp.PlainAuth("", conf.App.SMTPFrom, conf.App.SMTPPassword, host)
	if err := c.Auth(auth); err != nil {
		c.Close()
		return nil, err
	}
	return &Client{c: c, from: conf.App.SMTPFrom}, nil
}

// Send delivers a single message over the persistent connection.
func (cl *Client) Send(m *Message) error {
	if err := cl.c.Mail(cl.from); err != nil {
		return err
	}
	for _, recp := range m.Recipients() {
		if err := cl.c.Rcpt(recp); err != nil {
			return err
		}
	}
	w, err := cl.c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(m.Bytes()); err != nil {
		return err
	}
	return w.Close()
}

// Reset aborts the current mail transaction without closing the connection.
func (cl *Client) Reset() error {
	return cl.c.Reset()
}

// Close terminates the SMTP session and closes the connection.
func (cl *Client) Close() error {
	return cl.c.Quit()
}
