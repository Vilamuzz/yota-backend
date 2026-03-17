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

type RoleResponse struct {
	ID   int8   `json:"id"`
	Role string `json:"role"`
}

type UserListResponse struct {
	Users      []UserResponse       `json:"users"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

type RoleListResponse struct {
	Roles []RoleResponse `json:"roles"`
}

func (u *User) toUserResponse() UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Status:    u.Status,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func (u *User) toUserProfileResponse() UserProfileResponse {
	return UserProfileResponse{
		ID:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role.Role,
	}
}

func toUserListResponse(users []User, pagination pkg.CursorPagination) UserListResponse {
	var responses []UserResponse
	for _, user := range users {
		responses = append(responses, user.toUserResponse())
	}
	if responses == nil {
		responses = []UserResponse{}
	}
	return UserListResponse{
		Users:      responses,
		Pagination: pagination,
	}
}

func toRoleResponse(role Role) RoleResponse {
	return RoleResponse{
		ID:   int8(role.ID),
		Role: role.Role,
	}
}

func toRoleListResponse(roles []Role) RoleListResponse {
	var responses []RoleResponse
	for _, role := range roles {
		responses = append(responses, toRoleResponse(role))
	}
	if responses == nil {
		responses = []RoleResponse{}
	}
	return RoleListResponse{
		Roles: responses,
	}
}
