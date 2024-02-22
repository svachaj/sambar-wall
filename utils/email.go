package utils

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewEmailService(host string, port int, username string, password string) *EmailService {
	return &EmailService{Host: host, Port: port, Username: username, Password: password}
}

func (es *EmailService) SendEmail(subject string, body string, to string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(es.Host, es.Port, es.Username, es.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
