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

func (e *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("FE_URL"), resetToken)

	subject := "Password Reset Request"
	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>Password Reset Request</h2>
            <p>You have requested to reset your password. Click the link below to reset your password:</p>
            <p><a href="%s">Reset Password</a></p>
            <p>This link will expire in 1 hour.</p>
            <p>If you did not request this, please ignore this email.</p>
        </body>
        </html>
    `, resetURL)

	return e.SendEmail(to, subject, body)
}
