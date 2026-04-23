package account

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
)

type CreateAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserProfileRequest struct {
	Username             string                `json:"username" form:"username"`
	Email                string                `json:"email" form:"email"`
	DefaultAccountRoleID int                   `json:"default_account_role_id" form:"default_account_role_id"`
	Phone                string                `json:"phone" form:"phone"`
	Address              string                `json:"address" form:"address"`
	ProfilePicture       *multipart.FileHeader `json:"profile_picture" form:"profile_picture" swaggerignore:"true"`
}

type SetAccountBanStatusRequest struct {
	BanStatus bool `json:"ban_status"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AccountQueryParam struct {
	Search    string             `form:"search"`
	RoleID    int                `form:"role_id"`
	IsBanned  *bool              `form:"is_banned"`
	SortOrder enum.SortOrderEnum `form:"sort_order"`
	pkg.PaginationParams
}

type UpdateAccountRoleRequest struct {
	IsActive bool `json:"is_active"`
}
