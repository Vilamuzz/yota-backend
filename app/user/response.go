package user

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type UserProfileResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    bool      `json:"status"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UserListResponse struct {
	Users      []UserResponse       `json:"users"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
