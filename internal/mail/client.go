package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"

	"webgx/internal/conf"
)

// Client is a reusable SMTP client. It holds a single persistent connection so
// multiple messages can be sent without re-establishing the session each time.
type Client struct {
	c    *smtp.Client
	from string
}

// NewClient dials the configured SMTP server and authenticates. smtp.tls picks
// the transport: "ssl" is implicit TLS, "starttls" requires the STARTTLS
// upgrade, "none" stays plaintext, and "auto" uses implicit TLS on port 465 and
// opportunistic STARTTLS on every other port. Authentication runs only when a
// password is configured; the username falls back to the From address.
func NewClient() (*Client, error) {
	host, _, splitErr := net.SplitHostPort(conf.App.SMTPAddr)
	if splitErr != nil {
		return nil, splitErr
	}

	mode := conf.App.SMTPTLS
	if mode == conf.SMTPTLSAuto && conf.App.SMTPPort == 465 {
		mode = conf.SMTPTLSSSL
	}

	var c *smtp.Client
	if mode == conf.SMTPTLSSSL {
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
		if mode != conf.SMTPTLSNone {
			// Extension reports (supported, params); STARTTLS carries no params,
			// so the boolean is the whole answer.
			if ok, _ := c.Extension("STARTTLS"); ok {
				if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
					c.Close()
					return nil, err
				}
			} else if mode == conf.SMTPTLSStartTLS {
				c.Close()
				return nil, fmt.Errorf("smtp: %s does not advertise STARTTLS", host)
			}
		}
	}

	if conf.App.SMTPPassword != "" {
		user := conf.App.SMTPUsername
		if user == "" {
			user = conf.App.SMTPFrom
		}
		if user == "" {
			c.Close()
			return nil, errors.New("smtp: username or from is required when a password is set")
		}
		if err := c.Auth(smtp.PlainAuth("", user, conf.App.SMTPPassword, host)); err != nil {
			c.Close()
			return nil, err
		}
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
