package utils

import (
	"fmt"
	"net/smtp"
)

type EmailSender interface {
	SendEmail(to string, subject string, body string) error
}

type smtpEmailSender struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewEmailSender(host, port, user, pass, from string) EmailSender {
	return &smtpEmailSender{host, port, user, pass, from}
}

func (s *smtpEmailSender) SendEmail(to string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
