package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"

	"webtplmst/internal/conf"
)

// Part is a single MIME part of a Message (text, html, or attachment).
type Part struct {
	contentType string
	filename    string
	body        []byte
}

// Message is an RFC 822 email with support for plain text, HTML, attachments
// and extra recipients/headers.
type Message struct {
	from    string
	to      []string
	cc      []string
	bcc     []string
	subject string
	headers map[string]string
	parts   []Part
}

// NewMessage creates an empty message addressed to the given recipients.
// The From address is taken from the SMTP configuration.
func NewMessage(to []string, subject string) *Message {
	return &Message{
		from:    conf.App.SMTPFrom,
		to:      to,
		subject: subject,
		headers: make(map[string]string),
	}
}

// SetBodyText sets a plain text alternative body.
func (m *Message) SetBodyText(text string) *Message {
	m.parts = append(m.parts, Part{contentType: "text/plain; charset=UTF-8", body: []byte(text)})
	return m
}

// SetBodyHTML sets an HTML alternative body.
func (m *Message) SetBodyHTML(html string) *Message {
	m.parts = append(m.parts, Part{contentType: "text/html; charset=UTF-8", body: []byte(html)})
	return m
}

// Attach appends an attachment with the given filename, content type and bytes.
func (m *Message) Attach(filename, contentType string, body []byte) *Message {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	m.parts = append(m.parts, Part{contentType: contentType, filename: filename, body: body})
	return m
}

// CC adds carbon-copy recipients.
func (m *Message) CC(addrs ...string) *Message {
	m.cc = append(m.cc, addrs...)
	return m
}

// BCC adds blind carbon-copy recipients. They receive the mail but are not
// listed in the headers.
func (m *Message) BCC(addrs ...string) *Message {
	m.bcc = append(m.bcc, addrs...)
	return m
}

// SetHeader overrides or adds a message header (e.g. "Reply-To").
func (m *Message) SetHeader(name, value string) *Message {
	m.headers[name] = value
	return m
}

// Recipients returns the union of to, cc and bcc addresses used for the SMTP
// RCPT command.
func (m *Message) Recipients() []string {
	recps := make([]string, 0, len(m.to)+len(m.cc)+len(m.bcc))
	recps = append(recps, m.to...)
	recps = append(recps, m.cc...)
	recps = append(recps, m.bcc...)
	return recps
}

// Bytes renders the complete RFC 822 message, ready to be passed to the SMTP
// DATA command.
func (m *Message) Bytes() []byte {
	var buf bytes.Buffer

	buf.WriteString("From: ")
	buf.WriteString(m.from)
	buf.WriteString("\r\n")
	buf.WriteString("To: ")
	buf.WriteString(strings.Join(m.to, ", "))
	buf.WriteString("\r\n")
	if len(m.cc) > 0 {
		buf.WriteString("Cc: ")
		buf.WriteString(strings.Join(m.cc, ", "))
		buf.WriteString("\r\n")
	}
	buf.WriteString("Subject: ")
	buf.WriteString(mime.BEncoding.Encode("UTF-8", m.subject))
	buf.WriteString("\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")

	names := make([]string, 0, len(m.headers))
	for k := range m.headers {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(m.headers[k])
		buf.WriteString("\r\n")
	}

	var hasAttachment bool
	var hasBody bool
	for _, p := range m.parts {
		if p.filename != "" {
			hasAttachment = true
		} else {
			hasBody = true
		}
	}

	switch {
	case hasAttachment:
		m.writeMixed(&buf, hasBody)
	default:
		m.writeAlternative(&buf)
	}
	return buf.Bytes()
}

// writeAlternative emits a multipart/alternative (text + html) for messages
// with no attachments.
func (m *Message) writeAlternative(buf *bytes.Buffer) {
	mp := multipart.NewWriter(buf)
	buf.WriteString("Content-Type: multipart/alternative; boundary=")
	buf.WriteString(mp.Boundary())
	buf.WriteString("\r\n\r\n")
	for _, p := range m.parts {
		if p.filename != "" {
			continue
		}
		writePart(mp, p)
	}
	mp.Close()
}

// writeMixed emits a multipart/mixed message. The text/html bodies are nested
// in a multipart/alternative part; attachments follow as separate base64 parts.
func (m *Message) writeMixed(buf *bytes.Buffer, hasBody bool) {
	mp := multipart.NewWriter(buf)
	buf.WriteString("Content-Type: multipart/mixed; boundary=")
	buf.WriteString(mp.Boundary())
	buf.WriteString("\r\n\r\n")

	if hasBody {
		alt := multipart.NewWriter(buf)
		buf.WriteString("--")
		buf.WriteString(mp.Boundary())
		buf.WriteString("\r\n")
		buf.WriteString("Content-Type: multipart/alternative; boundary=")
		buf.WriteString(alt.Boundary())
		buf.WriteString("\r\n\r\n")
		for _, p := range m.parts {
			if p.filename != "" {
				continue
			}
			writePart(alt, p)
		}
		alt.Close()
	}

	for _, p := range m.parts {
		if p.filename == "" {
			continue
		}
		headers := textproto.MIMEHeader{
			"Content-Type":              {fmt.Sprintf("%s; name=%s", p.contentType, mime.BEncoding.Encode("UTF-8", p.filename))},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=%s", mime.BEncoding.Encode("UTF-8", p.filename))},
		}
		w, _ := mp.CreatePart(headers)
		enc := base64.NewEncoder(base64.StdEncoding, w)
		enc.Write(p.body)
		enc.Close()
	}
	mp.Close()
}

// writePart writes a single 7bit MIME part (text or html body) to a writer.
func writePart(mp *multipart.Writer, p Part) {
	w, _ := mp.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {p.contentType},
		"Content-Transfer-Encoding": {"7bit"},
	})
	w.Write(p.body)
}
