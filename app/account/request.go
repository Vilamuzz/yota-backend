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
	DefaultAccountRoleID int                   `json:"defaultAccountRoleId" form:"defaultAccountRoleId"`
	Phone                string                `json:"phone" form:"phone"`
	Address              string                `json:"address" form:"address"`
	ProfilePicture       *multipart.FileHeader `json:"profilePicture" form:"profilePicture" swaggerignore:"true"`
}

type SetAccountBanStatusRequest struct {
	BanStatus bool `json:"banStatus"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type AccountQueryParam struct {
	Search    string             `form:"search"`
	RoleID    int                `form:"roleId"`
	IsBanned  *bool              `form:"isBanned"`
	SortOrder enum.SortOrderEnum `form:"sortOrder"`
	pkg.PaginationParams
}

type UpdateAccountRoleRequest struct {
	IsActive bool `json:"isActive"`
}
