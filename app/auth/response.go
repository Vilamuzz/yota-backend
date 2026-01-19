package auth

import "github.com/Vilamuzz/yota-backend/app/user"

type LoginResponse struct {
	Token string           `json:"token"`
	User  user.UserProfile `json:"user"`
}
