package provider

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
)

type SMTPEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPEmailSender(host, port, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailSender) SendNotification(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.from, to, subject, body)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		log.Printf("[SMTPProvider] Failed to send email to %s: %v", to, err)
		return fmt.Errorf("smtp send failed: %w", err)
	}

	log.Printf("[SMTPProvider] Email sent to %s | Subject: %s", to, subject)
	return nil
}
