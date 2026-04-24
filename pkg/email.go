package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	if e.config.APIKey != "" {
		url := "https://api.resend.com/emails"

		payload := map[string]interface{}{
			"from":    e.config.From,
			"to":      []string{to},
			"subject": subject,
			"html":    body,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+e.config.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			var errResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errResp)
			return fmt.Errorf("resend api error: %v", errResp)
		}

		return nil
	}

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
