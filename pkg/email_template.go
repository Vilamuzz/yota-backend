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

// EmailVerificationTemplate generates the HTML body for email verification.
func EmailVerificationTemplate(recipientName, verificationURL string) string {
	return fmt.Sprintf(`
        <html>
        <body>
            <h2>Email Verification</h2>
            <p>Hi %s,</p>
            <p>Thank you for registering! Please verify your email address by clicking the link below:</p>
            <p><a href="%s">Verify Email</a></p>
            <p>This link will expire in 24 hours.</p>
            <p>If you did not create an account, please ignore this email.</p>
        </body>
        </html>
    `, recipientName, verificationURL)
}
