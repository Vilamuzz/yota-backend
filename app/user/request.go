package user

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	RoleID int8  `json:"role_id"`
	Status *bool `json:"status"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserQueryParam struct {
	pkg.PaginationParams
	Search string `form:"search"`
	Role   int8   `form:"role"`
	Status *bool  `form:"status"`
}
