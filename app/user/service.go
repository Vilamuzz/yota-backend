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
	ListUsers(ctx context.Context, queryParams UserQueryParam) pkg.Response
	GetUserByID(ctx context.Context, userID string) pkg.Response
	GetProfile(ctx context.Context, userID string) pkg.Response
	UpdateUser(ctx context.Context, userID string, payload UpdateUserRequest) pkg.Response
	UpdateProfile(ctx context.Context, userID string, payload UpdateProfileRequest) pkg.Response
	UpdatePassword(ctx context.Context, userID string, payload UpdatePasswordRequest) pkg.Response
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

	// Map to RoleResponse
	var roleResponses []RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, RoleResponse{
			ID:   role.ID,
			Role: role.Role,
		})
	}

	return pkg.NewResponse(http.StatusOK, "Roles retrieved successfully", nil, roleResponses)
}

func (s *service) ListUsers(ctx context.Context, queryParams UserQueryParam) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	usingPrevCursor := queryParams.PrevCursor != ""

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.Role != 0 {
		options["role_id"] = queryParams.Role
	}
	if queryParams.Status != nil {
		options["status"] = *queryParams.Status
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
	}
	if queryParams.NextCursor != "" {
		options["next_cursor"] = queryParams.NextCursor
	}
	if usingPrevCursor {
		options["prev_cursor"] = queryParams.PrevCursor
	}

	users, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to retrieve users", nil, nil)
	}

	// When traversing backwards the repo returns ASC order; check overflow then reverse
	hasMore := len(users) > queryParams.Limit
	if hasMore {
		users = users[:queryParams.Limit]
	}

	// Reverse the slice when using prev_cursor so the page is in DESC order
	if usingPrevCursor {
		for i, j := 0, len(users)-1; i < j; i, j = i+1, j-1 {
			users[i], users[j] = users[j], users[i]
		}
	}

	// Map to UserResponse
	var userProfiles []UserResponse
	for _, user := range users {
		userProfiles = append(userProfiles, UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			Status:    user.Status,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		})
	}

	var nextCursor, prevCursor string

	hasNext := (!usingPrevCursor && hasMore) || (usingPrevCursor && queryParams.NextCursor == "")
	hasPrev := (usingPrevCursor && hasMore) || (!usingPrevCursor && queryParams.NextCursor != "")

	if len(userProfiles) > 0 {
		first := userProfiles[0]
		last := userProfiles[len(userProfiles)-1]

		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID)
		}
	}

	resData := UserListResponse{
		Users: userProfiles,
		Pagination: pkg.CursorPagination{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			Limit:      queryParams.Limit,
		},
	}

	return pkg.NewResponse(http.StatusOK, "Users list retrieved successfully", nil, resData)
}

func (s *service) GetUserByID(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FindOne(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	// Map to UserProfile response
	userProfile := UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	return pkg.NewResponse(http.StatusOK, "User details retrieved successfully", nil, userProfile)
}

func (s *service) GetProfile(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FindOne(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	// Map to UserProfile response
	userProfile := UserProfileResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role.Role,
	}

	return pkg.NewResponse(http.StatusOK, "User profile retrieved successfully", nil, userProfile)
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

	updateMap := make(map[string]interface{})
	if payload.RoleID != 0 {
		updateMap["role_id"] = payload.RoleID
	}

	if payload.Status != nil {
		updateMap["status"] = *payload.Status
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}
	if len(updateMap) == 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"update_data": "No fields to update"}, nil)
	}

	err := s.repo.Update(ctx, userID, updateMap)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update user", nil, nil)
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
	err = s.repo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password updated successfully", nil, nil)
}
