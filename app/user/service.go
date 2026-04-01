package user

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	ListRoles(ctx context.Context) pkg.Response
	ListUsers(ctx context.Context, params UserQueryParam) pkg.Response
	GetUserByID(ctx context.Context, id string) pkg.Response
	GetProfile(ctx context.Context, id string) pkg.Response
	UpdateUser(ctx context.Context, id string, payload UpdateUserRequest) pkg.Response
	UpdateProfile(ctx context.Context, id string, payload UpdateProfileRequest) pkg.Response
	UpdatePassword(ctx context.Context, id string, payload UpdatePasswordRequest) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(r Repository, timeout time.Duration) Service {
	return &service{
		repo:    r,
		timeout: timeout,
	}
}

func (s *service) ListRoles(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	roles, err := s.repo.FindAllRoles(ctx)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to retrieve roles", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Roles retrieved successfully", nil, toRoleListResponse(roles))
}

func (s *service) ListUsers(ctx context.Context, params UserQueryParam) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	usingPrevCursor := params.PrevCursor != ""

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.Role != 0 {
		options["role_id"] = params.Role
	}
	if params.Status != nil {
		options["status"] = *params.Status
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = params.PrevCursor
	}

	users, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to retrieve users", nil, nil)
	}

	hasMore := len(users) > params.Limit
	if hasMore {
		users = users[:params.Limit]
	}
	if usingPrevCursor {
		for i, j := 0, len(users)-1; i < j; i, j = i+1, j-1 {
			users[i], users[j] = users[j], users[i]
		}
	}

	var nextCursor, prevCursor string
	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && params.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && params.NextCursor != "")

	if len(users) > 0 {
		first := users[0]
		last := users[len(users)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Users list retrieved successfully", nil, toUserListResponse(users, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      params.Limit,
	}))
}

func (s *service) GetUserByID(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FindOne(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "User details retrieved successfully", nil, user.toUserResponse())
}

func (s *service) GetProfile(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FindOne(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "User profile retrieved successfully", nil, user.toUserProfileResponse())
}

func (s *service) UpdateUser(ctx context.Context, userID string, payload UpdateUserRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.RoleID != 0 {
		_, err := s.repo.FindRoleByID(ctx, payload.RoleID)
		if err != nil {
			errValidation["role"] = "Invalid role"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if payload.RoleID == 0 && payload.Status == nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	// Update roles via the join table if provided
	if payload.RoleID != 0 {
		if err := s.repo.UpdateUserRoles(ctx, userID, payload.RoleID); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to update user role", nil, nil)
		}
	}

	// Update direct user columns (status, etc.) if provided
	if payload.Status != nil {
		if err := s.repo.Update(ctx, userID, map[string]interface{}{"status": *payload.Status}); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to update user", nil, nil)
		}
	}

	return pkg.NewResponse(http.StatusOK, "User updated successfully", nil, nil)
}

func (s *service) UpdateProfile(ctx context.Context, userID string, payload UpdateProfileRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	updateMap := make(map[string]interface{})
	if payload.Username != "" {
		updateMap["username"] = payload.Username
	}
	if payload.Email != "" {
		updateMap["email"] = payload.Email
	}
	if len(updateMap) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	err := s.repo.Update(ctx, userID, updateMap)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update profile", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Profile updated successfully", nil, nil)
}

func (s *service) UpdatePassword(ctx context.Context, userID string, payload UpdatePasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.CurrentPassword == "" {
		errValidation["current_password"] = "Current password is required"
	}
	if payload.NewPassword == "" {
		errValidation["new_password"] = "New password is required"
	}
	if !pkg.IsValidLengthPassword(payload.NewPassword) {
		errValidation["new_password"] = "New password must be at least 8 characters long"
	}
	if !pkg.IsStrongPassword(payload.NewPassword) {
		errValidation["new_password"] = "New password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	user, err := s.repo.FindOne(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.CurrentPassword)); err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Current password is incorrect", nil, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash new password", nil, nil)
	}

	err = s.repo.Update(ctx, userID, map[string]interface{}{"password": string(hashedPassword)})
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Password updated successfully", nil, nil)
}
