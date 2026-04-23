package account

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type UserProfileResponse struct {
	ID             string                 `json:"id"`
	Username       string                 `json:"username"`
	Email          string                 `json:"email"`
	Roles          []AccountRolesResponse `json:"roles"`
	ProfilePicture string                 `json:"profile_picture"`
	Phone          string                 `json:"phone"`
	Address        string                 `json:"address"`
}

type AccountResponse struct {
	ID        string                 `json:"id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	IsBanned  bool                   `json:"is_banned"`
	Roles     []AccountRolesResponse `json:"roles"`
	CreatedAt time.Time              `json:"created_at"`
}

type AccountListResponse struct {
	Accounts   []AccountResponse    `json:"accounts"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

type AccountRolesResponse struct {
	RoleID    int    `json:"role_id"`
	RoleName  string `json:"role_name"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
}

type RoleResponse struct {
	ID       int    `json:"id"`
	RoleName string `json:"role_name"`
}

type RolesResponse struct {
	Roles []RoleResponse `json:"roles"`
}

func (a *Account) toAccountResponse() AccountResponse {
	return AccountResponse{
		ID:        a.ID.String(),
		Username:  a.UserProfile.Username,
		Email:     a.Email,
		IsBanned:  a.IsBanned,
		Roles:     toAccountRolesResponse(a.AccountRoles),
		CreatedAt: a.CreatedAt,
	}
}

func (a *Account) toUserProfileResponse() UserProfileResponse {
	return UserProfileResponse{
		ID:             a.ID.String(),
		Username:       a.UserProfile.Username,
		Email:          a.Email,
		Roles:          toAccountRolesResponse(a.AccountRoles),
		Phone:          a.UserProfile.Phone,
		Address:        a.UserProfile.Address,
		ProfilePicture: a.UserProfile.ProfilePicture,
	}
}

func toAccountRolesResponse(accountRoles []AccountRole) []AccountRolesResponse {
	if len(accountRoles) == 0 {
		return []AccountRolesResponse{}
	}

	responses := make([]AccountRolesResponse, 0, len(accountRoles))
	for _, accountRole := range accountRoles {
		roleName := string(accountRole.Role.Name)
		responses = append(responses, AccountRolesResponse{
			RoleID:    accountRole.RoleID,
			RoleName:  roleName,
			IsDefault: accountRole.IsDefault,
			IsActive:  accountRole.IsActive,
		})
	}

	return responses
}

func toAccountListResponse(accounts []Account, pagination pkg.CursorPagination) AccountListResponse {
	var responses []AccountResponse
	for _, account := range accounts {
		responses = append(responses, account.toAccountResponse())
	}
	if responses == nil {
		responses = []AccountResponse{}
	}
	return AccountListResponse{
		Accounts:   responses,
		Pagination: pagination,
	}
}

func toRolesResponse(roles []Role) RolesResponse {
	var responses []RoleResponse
	for _, role := range roles {
		responses = append(responses, RoleResponse{
			ID:       role.ID,
			RoleName: string(role.Name),
		})
	}
	if responses == nil {
		responses = []RoleResponse{}
	}
	return RolesResponse{
		Roles: responses,
	}
}
