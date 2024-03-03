package utils

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type IEmailService interface {
	SendEmail(subject string, body string, to string) error
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

type MockEmailService struct {
}

func NewMockEmailService() IEmailService {
	return &MockEmailService{}
}

func (es *MockEmailService) SendEmail(subject string, body string, to string) error {
	return nil
}
