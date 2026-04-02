package user

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	DefaultRoleID int8   `json:"default_role_id"`
	Roles         []int8 `json:"roles"`
	Status        *bool  `json:"status"`
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
