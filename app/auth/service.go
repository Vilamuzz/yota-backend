package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) pkg.Response
	Login(ctx context.Context, req LoginRequest) pkg.Response
	ForgetPassword(ctx context.Context, req ForgetPasswordRequest) pkg.Response
	ResetPassword(ctx context.Context, req ResetPasswordRequest) pkg.Response
}

type service struct {
	userRepo       user.Repository
	resetTokenRepo Repository
	emailService   *pkg.EmailService
	contextTimeout time.Duration
}

func NewService(userRepo user.Repository, resetTokenRepo Repository, timeout time.Duration) Service {
	return &service{
		userRepo:       userRepo,
		resetTokenRepo: resetTokenRepo,
		emailService:   pkg.NewEmailService(),
		contextTimeout: timeout,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate email format
	if !pkg.IsValidEmail(req.Email) {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid email format", nil, nil)
	}

	// Validate password length
	if !pkg.IsValidLengthPassword(req.Password) {
		return pkg.NewResponse(http.StatusBadRequest, "Password must be at least 8 characters", nil, nil)
	}

	// Validate password strength
	if !pkg.IsStrongPassword(req.Password) {
		return pkg.NewResponse(http.StatusBadRequest, "Password must contain uppercase, lowercase, and number", nil, nil)
	}

	// Check if email already exists
	existingUser, _ := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if existingUser != nil {
		return pkg.NewResponse(http.StatusConflict, "Email already registered", nil, nil)
	}

	// Check if username already exists
	existingUser, _ = s.userRepo.FetchOneUser(ctx, map[string]interface{}{"username": req.Username})
	if existingUser != nil {
		return pkg.NewResponse(http.StatusConflict, "Username already taken", nil, nil)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	// Create user
	newUser := &user.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      user.RoleUser,
		Status:    user.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateOneUser(ctx, newUser); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "User registered successfully", nil, nil)
}

func (s *service) Login(ctx context.Context, req LoginRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Find user by email
	existingUser, err := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	// Check if user is banned
	if existingUser.Status == user.StatusBanned {
		return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	// Generate JWT token with role
	ttl := config.GetJWTTTL()
	claims := &jwt_pkg.UserJWTClaims{
		UserID: existingUser.ID.String(),
		Role:   string(existingUser.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt_pkg.GenerateJWTToken(claims)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate token", nil, nil)
	}

	loginResponse := LoginResponse{
		Token: token,
		User: user.UserProfile{
			ID:       existingUser.ID.String(),
			Username: existingUser.Username,
			Email:    existingUser.Email,
			Role:     string(existingUser.Role),
		},
	}

	return pkg.NewResponse(http.StatusOK, "Login successful", nil, loginResponse)
}

func (s *service) ForgetPassword(ctx context.Context, req ForgetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Find user by email
	existingUser, err := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if err != nil {
		// Return success even if user not found (security best practice)
		return pkg.NewResponse(http.StatusOK, "If the email exists, a reset link has been sent", nil, nil)
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate reset token", nil, nil)
	}
	resetToken := hex.EncodeToString(tokenBytes)

	// Create password reset token record
	passwordReset := &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    existingUser.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := s.resetTokenRepo.CreatePasswordResetToken(ctx, passwordReset); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create reset token", nil, nil)
	}

	// Send email
	if err := s.emailService.SendPasswordResetEmail(req.Email, resetToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send reset email", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset email sent successfully", nil, nil)
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate new password
	if !pkg.IsValidLengthPassword(req.NewPassword) {
		return pkg.NewResponse(http.StatusBadRequest, "Password must be at least 8 characters", nil, nil)
	}

	if !pkg.IsStrongPassword(req.NewPassword) {
		return pkg.NewResponse(http.StatusBadRequest, "Password must contain uppercase, lowercase, and number", nil, nil)
	}

	// Fetch reset token
	resetToken, err := s.resetTokenRepo.FetchPasswordResetToken(ctx, req.Token)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid or expired reset token", nil, nil)
	}

	// Check if token is used
	if resetToken.Used {
		return pkg.NewResponse(http.StatusBadRequest, "Reset token already used", nil, nil)
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return pkg.NewResponse(http.StatusBadRequest, "Reset token has expired", nil, nil)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	// Update user password
	if err := s.userRepo.UpdateUserPassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}

	// Mark token as used
	resetToken.Used = true
	if err := s.resetTokenRepo.UpdatePasswordResetToken(ctx, resetToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update token status", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset successfully", nil, nil)
}
