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
	"github.com/markbates/goth"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) pkg.Response
	Login(ctx context.Context, req LoginRequest) pkg.Response
	ForgetPassword(ctx context.Context, req ForgetPasswordRequest) pkg.Response
	ResetPassword(ctx context.Context, req ResetPasswordRequest) pkg.Response
	OAuthLogin(ctx context.Context, provider string, gothUser goth.User) pkg.Response
	VerifyEmail(ctx context.Context, token string) pkg.Response
	ResendVerificationEmail(ctx context.Context, email string) pkg.Response
}

type service struct {
	userRepo       user.Repository
	resetTokenRepo Repository
	emailService   *pkg.EmailService
	contextTimeout time.Duration
}

func NewService(resetTokenRepo Repository, userRepo user.Repository, timeout time.Duration) Service {
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

	errValidation := make(map[string]string)
	if req.Email == "" {
		errValidation["email"] = "Email is required"
	}
	if req.Username == "" {
		errValidation["username"] = "Username is required"
	}
	if req.Password == "" {
		errValidation["password"] = "Password is required"
	}
	// Validate email format
	if !pkg.IsValidEmail(req.Email) {
		errValidation["email"] = "Invalid email format"
	}

	// Validate password length
	if !pkg.IsValidLengthPassword(req.Password) {
		errValidation["password"] = "Password must be at least 8 characters"
	}

	// Validate password strength
	if !pkg.IsStrongPassword(req.Password) {
		errValidation["password"] = "Password must contain uppercase, lowercase, and number"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
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

	// Create user with email_verified = false
	newUser := &user.User{
		ID:            uuid.New(),
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Role:          user.RoleUser,
		Status:        true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.CreateOneUser(ctx, newUser); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
	}

	// Generate email verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate verification token", nil, nil)
	}
	verificationToken := hex.EncodeToString(tokenBytes)

	// Create email verification token record
	emailVerification := &EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    newUser.ID,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := s.resetTokenRepo.CreateEmailVerificationToken(ctx, emailVerification); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create verification token", nil, nil)
	}

	// Send verification email
	if err := s.emailService.SendEmailVerification(newUser.Email, newUser.Username, verificationToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send verification email", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Registration successful. Please check your email to verify your account.", nil, map[string]interface{}{
		"email": newUser.Email,
	})
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
	if !existingUser.Status {
		return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	// Check if email is verified
	if !existingUser.EmailVerified {
		return pkg.NewResponse(http.StatusForbidden, "Please verify your email before logging in", nil, nil)
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

	loginResponse := AuthResponse{
		Token: token,
	}

	return pkg.NewResponse(http.StatusOK, "Login successful", nil, loginResponse)
}

func (s *service) VerifyEmail(ctx context.Context, token string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Fetch verification token
	errValidation := make(map[string]string)
	verificationToken, err := s.resetTokenRepo.FetchEmailVerificationToken(ctx, token)
	if err != nil {
		errValidation["token"] = "Invalid or expired verification token"
	}

	if verificationToken.Used {
		errValidation["token"] = "Verification token already used"
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		errValidation["token"] = "Verification token has expired"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Verify user email
	if err := s.userRepo.VerifyUserEmail(ctx, verificationToken.UserID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to verify email", nil, nil)
	}

	// Mark token as used
	verificationToken.Used = true
	if err := s.resetTokenRepo.UpdateEmailVerificationToken(ctx, verificationToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update token status", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Email verified successfully", nil, nil)
}

func (s *service) ResendVerificationEmail(ctx context.Context, email string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Find user by email
	existingUser, err := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": email})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	// Check if already verified
	if existingUser.EmailVerified {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"email": "Email already verified"}, nil)
	}

	// Generate new verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate verification token", nil, nil)
	}
	verificationToken := hex.EncodeToString(tokenBytes)

	// Create email verification token record
	emailVerification := &EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    existingUser.ID,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := s.resetTokenRepo.CreateEmailVerificationToken(ctx, emailVerification); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create verification token", nil, nil)
	}

	// Send verification email
	if err := s.emailService.SendEmailVerification(existingUser.Email, existingUser.Username, verificationToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send verification email", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Verification email sent successfully", nil, nil)
}

func (s *service) ForgetPassword(ctx context.Context, req ForgetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Find user by email
	existingUser, err := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if err != nil {
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
	if err := s.emailService.SendPasswordResetEmail(req.Email, existingUser.Username, resetToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send reset email", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset email sent successfully", nil, nil)
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate new password
	errValidation := make(map[string]string)
	if req.NewPassword == "" {
		errValidation["new_password"] = "New password is required"
	}

	if !pkg.IsValidLengthPassword(req.NewPassword) {
		errValidation["new_password"] = "Password must be at least 8 characters"
	}

	if !pkg.IsStrongPassword(req.NewPassword) {
		errValidation["new_password"] = "Password must contain uppercase, lowercase, and number"
	}

	// Fetch reset token
	resetToken, err := s.resetTokenRepo.FetchPasswordResetToken(ctx, req.Token)
	if err != nil {
		errValidation["token"] = "Invalid or expired reset token"
	}

	// Check if token is used
	if resetToken.Used {
		errValidation["token"] = "Reset token already used"
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		errValidation["token"] = "Reset token has expired"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
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

func (s *service) OAuthLogin(ctx context.Context, provider string, gothUser goth.User) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Try to find existing user by email
	existingUser, err := s.userRepo.FetchOneUser(ctx, map[string]interface{}{"email": gothUser.Email})

	var currentUser *user.User

	if err != nil {
		// User doesn't exist, create new user
		username := gothUser.NickName
		if username == "" {
			username = gothUser.Email
		}

		newUser := &user.User{
			ID:        uuid.New(),
			Username:  username,
			Email:     gothUser.Email,
			Password:  "",
			Role:      user.RoleUser,
			Status:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.userRepo.CreateOneUser(ctx, newUser); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
		}

		currentUser = newUser
	} else {
		// User exists
		if !existingUser.Status {
			return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
		}
		currentUser = existingUser
	}

	// Generate JWT token
	ttl := config.GetJWTTTL()
	claims := &jwt_pkg.UserJWTClaims{
		UserID: currentUser.ID.String(),
		Role:   string(currentUser.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt_pkg.GenerateJWTToken(claims)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate token", nil, nil)
	}

	authResponse := AuthResponse{
		Token: token,
	}

	return pkg.NewResponse(http.StatusOK, "OAuth login successful", nil, authResponse)
}
