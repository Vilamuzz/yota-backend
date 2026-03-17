package user
import "github.com/Vilamuzz/yota-backend/pkg"

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	RoleID int8  `json:"role_id" binding:"omitempty"`
	Status *bool `json:"status" binding:"omitempty"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type UserQueryParam struct {
	pkg.PaginationParams
	Search     string `form:"search"`
	Role       int8   `form:"role"`
	Status     *bool  `form:"status"`
}
