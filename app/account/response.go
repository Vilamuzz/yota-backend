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
	ProfilePicture string                 `json:"profilePicture"`
	Phone          string                 `json:"phone"`
	Address        string                 `json:"address"`
}

type AccountResponse struct {
	ID        string                 `json:"id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	IsBanned  bool                   `json:"isBanned"`
	Roles     []AccountRolesResponse `json:"roles"`
	CreatedAt time.Time              `json:"createdAt"`
}

type AccountListResponse struct {
	Accounts   []AccountResponse    `json:"accounts"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

type AccountRolesResponse struct {
	RoleID    int    `json:"roleId"`
	RoleName  string `json:"roleName"`
	IsDefault bool   `json:"isDefault"`
	IsActive  bool   `json:"isActive"`
}

type RoleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
			ID:   role.ID,
			Name: string(role.Name),
		})
	}
	if responses == nil {
		responses = []RoleResponse{}
	}
	return RolesResponse{
		Roles: responses,
	}
}
