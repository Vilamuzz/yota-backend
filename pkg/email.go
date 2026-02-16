package pkg

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/Vilamuzz/yota-backend/config"
)

type EmailService struct {
	config config.SMTPConfig
}

func NewEmailService() *EmailService {
	return &EmailService{
		config: config.GetSMTPConfig(),
	}
}

func (e *EmailService) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
		"\r\n"+
		"%s\r\n", e.config.From, to, subject, body))

	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	return smtp.SendMail(addr, auth, e.config.From, []string{to}, msg)
}

func (e *EmailService) SendPasswordResetEmail(to, username, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("FE_URL"), resetToken)

	subject := "Password Reset Request"
	body := PasswordResetTemplate(username, resetURL)

	return e.SendEmail(to, subject, body)
}

func (e *EmailService) SendEmailVerification(to, username, verificationToken string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", os.Getenv("FE_URL"), verificationToken)

	subject := "Email Verification"
	body := EmailVerificationTemplate(username, verificationURL)

	return e.SendEmail(to, subject, body)
}
