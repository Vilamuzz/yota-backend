package auth

type AuthResponse struct {
	Token                 string `json:"token"`
	RequiresPasswordSetup bool   `json:"requiresPasswordSetup"`
}
