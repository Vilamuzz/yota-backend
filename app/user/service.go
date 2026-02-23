package user

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetUsersList(ctx context.Context, queryParams UserQueryParam) pkg.Response
	GetUserDetail(ctx context.Context, userID string) pkg.Response
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

func (s *service) GetUsersList(ctx context.Context, queryParams UserQueryParam) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if queryParams.Limit == 0 {
		queryParams.Limit = 10
	}

	options := map[string]interface{}{
		"limit": queryParams.Limit,
	}

	if queryParams.Role != 0 {
		options["role"] = queryParams.Role
	}
	if queryParams.Status != nil {
		options["status"] = *queryParams.Status
	}
	if queryParams.Search != "" {
		options["search"] = queryParams.Search
	}
	if queryParams.Cursor != "" {
		options["cursor"] = queryParams.Cursor
	}

	users, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to retrieve users", nil, nil)
	}
	// Map to UserProfile responses
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
	return pkg.NewResponse(http.StatusOK, "Users list retrieved successfully", nil, userProfiles)
}

func (s *service) GetUserDetail(ctx context.Context, userID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.repo.FindByID(ctx, userID)
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

func (s *service) UpdateUser(ctx context.Context, userID string, payload UpdateUserRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.Role != 0 {
		_, err := s.repo.FindRoleByID(ctx, payload.Role)
		if err != nil {
			errValidation["role"] = "Invalid role"
		}
	}

	updateMap := make(map[string]interface{})
	if payload.Role != 0 {
		updateMap["role"] = payload.Role
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

	err := s.repo.UpdateUser(ctx, userID, updateMap)
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

	err := s.repo.UpdateUser(ctx, userID, updateMap)
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

	user, err := s.repo.FindByID(ctx, userID)
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
	err = s.repo.UpdateUserPassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password updated successfully", nil, nil)
}
