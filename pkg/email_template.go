package pkg

import "fmt"

// PasswordResetTemplate generates the HTML body for a password reset email.
func PasswordResetTemplate(recipientName, resetURL string) string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hi %s,</p>
			<p>You have requested to reset your password. Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>
	`, recipientName, resetURL)
}
