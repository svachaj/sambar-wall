package utils

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type IEmailService interface {
	SendEmail(subject string, body string, to string) error
	SendEmailWithImage(subject string, body string, to string, image []byte, imageCID string) error
}
type EmailService struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewEmailService(host string, port int, username string, password string) IEmailService {
	return &EmailService{Host: host, Port: port, Username: username, Password: password}
}

var (
	styleScriptRe = regexp.MustCompile(`(?is)<(style|script)[^>]*>.*?</(style|script)>`)
	brTagRe       = regexp.MustCompile(`(?i)<br\s*/?>`)
	blockCloseRe  = regexp.MustCompile(`(?i)</(p|div|li|h[1-6])>`)
	htmlTagRe     = regexp.MustCompile(`<[^>]*>`)
	blankLinesRe  = regexp.MustCompile(`\n{3,}`)
)

// emailDomain derives the domain to use for the Message-ID and SMTP HELO
// from the SMTP account address, keeping From/HELO/Message-ID consistent
// without requiring a separate config value.
func emailDomain(address string) string {
	parts := strings.SplitN(address, "@", 2)
	if len(parts) == 2 && parts[1] != "" {
		return parts[1]
	}
	log.Warn().Str("username", address).Msg("email username has no domain part, falling back to localhost HELO/Message-ID domain")
	return "localhost"
}

func newMessageID(domain string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("<%d@%v>", time.Now().UnixNano(), domain)
	}
	return fmt.Sprintf("<%x@%v>", b, domain)
}

// plainTextFromHTML derives a text/plain fallback from an HTML fragment so
// the outgoing message can be sent as multipart/alternative.
func plainTextFromHTML(body string) string {
	text := styleScriptRe.ReplaceAllString(body, "")
	text = brTagRe.ReplaceAllString(text, "\n")
	text = blockCloseRe.ReplaceAllString(text, "\n")
	text = htmlTagRe.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = blankLinesRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

// wrapHTMLDocument ensures the body is a fully valid HTML document, since
// some mail filters penalize an HTML part that isn't wrapped in <html>/<body>.
func wrapHTMLDocument(body string) string {
	if strings.Contains(strings.ToLower(body), "<html") {
		return body
	}
	return fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>%v</body></html>`, body)
}

func (es *EmailService) domain() string {
	return emailDomain(es.Username)
}

func (es *EmailService) buildMessage(subject string, body string, to string) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", es.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetHeader("Message-ID", newMessageID(es.domain()))
	m.SetBody("text/plain", plainTextFromHTML(body))
	m.AddAlternative("text/html", wrapHTMLDocument(body))
	return m
}

func (es *EmailService) newDialer() *gomail.Dialer {
	d := gomail.NewDialer(es.Host, es.Port, es.Username, es.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	d.LocalName = es.domain()
	return d
}

func (es *EmailService) SendEmail(subject string, body string, to string) error {
	m := es.buildMessage(subject, body, to)
	return es.newDialer().DialAndSend(m)
}

func (es *EmailService) SendEmailWithImage(subject string, body string, to string, image []byte, imageCID string) error {
	m := es.buildMessage(subject, body, to)

	m.Embed(imageCID, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(image)
		return err
	}))

	return es.newDialer().DialAndSend(m)
}

type MockEmailService struct {
}

func NewMockEmailService() IEmailService {
	return &MockEmailService{}
}

func (es *MockEmailService) SendEmail(subject string, body string, to string) error {
	return nil
}

func (es *MockEmailService) SendEmailWithImage(subject string, body string, to string, image []byte, imageCID string) error {
	return nil
}
