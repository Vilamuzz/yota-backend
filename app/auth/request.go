package auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type OAuthCallbackRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type SwitchRoleRequest struct {
	Role string `json:"role"`
}
