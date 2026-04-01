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
	SwitchRole(ctx context.Context, claims jwt_pkg.UserJWTClaims, req SwitchRoleRequest) pkg.Response
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
	if !pkg.IsValidEmail(req.Email) {
		errValidation["email"] = "Invalid email format"
	}
	if !pkg.IsValidLengthPassword(req.Password) {
		errValidation["password"] = "Password must be at least 8 characters"
	}
	if !pkg.IsStrongPassword(req.Password) {
		errValidation["password"] = "Password must contain uppercase, lowercase, and number"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	existingUser, _ := s.userRepo.FindOne(ctx, map[string]interface{}{"email": req.Email})
	if existingUser != nil {
		return pkg.NewResponse(http.StatusConflict, "Email already registered", nil, nil)
	}

	existingUser, _ = s.userRepo.FindOne(ctx, map[string]interface{}{"username": req.Username})
	if existingUser != nil {
		return pkg.NewResponse(http.StatusConflict, "Username already taken", nil, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	newUser := &user.User{
		ID:            uuid.New(),
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Status:        true,
		EmailVerified: false,
		DefaultRoleID: 1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	newRoleUser := &user.UserRole{
		UserID: newUser.ID,
		RoleID: 1,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
	}

	if err := s.userRepo.CreateUserRole(ctx, newRoleUser); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user role", nil, nil)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate verification token", nil, nil)
	}
	verificationToken := hex.EncodeToString(tokenBytes)

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

	existingUser, err := s.userRepo.FindOne(ctx, map[string]interface{}{"email": req.Email})

	if err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	if !existingUser.Status {
		return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	if !existingUser.EmailVerified {
		return pkg.NewResponse(http.StatusForbidden, "Please verify your email before logging in", nil, nil)
	}

	if len(existingUser.Roles) == 0 {
		return pkg.NewResponse(http.StatusForbidden, "User has no assigned roles", nil, nil)
	}

	userRoles := make([]string, len(existingUser.Roles))
	for i, role := range existingUser.Roles {
		userRoles[i] = role.Role
	}

	ttl := config.GetJWTTTL()
	claims := &jwt_pkg.UserJWTClaims{
		UserID:     existingUser.ID.String(),
		Role:       userRoles,
		ActiveRole: existingUser.DefaultRole.Role,
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

	verificationToken, err := s.resetTokenRepo.FetchEmailVerificationToken(ctx, token)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"token": "Invalid or expired verification token"}, nil)
	}

	errValidation := make(map[string]string)
	if verificationToken.Used {
		errValidation["token"] = "Verification token already used"
	}

	if time.Now().After(verificationToken.ExpiresAt) {
		errValidation["token"] = "Verification token has expired"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if err := s.userRepo.VerifyEmail(ctx, verificationToken.UserID.String()); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to verify email", nil, nil)
	}

	verificationToken.Used = true
	if err := s.resetTokenRepo.UpdateEmailVerificationToken(ctx, verificationToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update token status", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Email verified successfully", nil, nil)
}

func (s *service) ResendVerificationEmail(ctx context.Context, email string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	existingUser, err := s.userRepo.FindOne(ctx, map[string]interface{}{"email": email})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "User not found", nil, nil)
	}

	if existingUser.EmailVerified {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"email": "Email already verified"}, nil)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate verification token", nil, nil)
	}
	verificationToken := hex.EncodeToString(tokenBytes)

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

	if err := s.emailService.SendEmailVerification(existingUser.Email, existingUser.Username, verificationToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send verification email", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Verification email sent successfully", nil, nil)
}

func (s *service) ForgetPassword(ctx context.Context, req ForgetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	existingUser, err := s.userRepo.FindOne(ctx, map[string]interface{}{"email": req.Email})
	if err != nil {
		return pkg.NewResponse(http.StatusOK, "If the email exists, a reset link has been sent", nil, nil)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate reset token", nil, nil)
	}
	resetToken := hex.EncodeToString(tokenBytes)

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

	if err := s.emailService.SendPasswordResetEmail(req.Email, existingUser.Username, resetToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to send reset email", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset email sent successfully", nil, nil)
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

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

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	resetToken, err := s.resetTokenRepo.FetchPasswordResetToken(ctx, req.Token)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"token": "Invalid or expired reset token"}, nil)
	}

	if resetToken.Used {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"token": "Reset token already used"}, nil)
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"token": "Reset token has expired"}, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	if err := s.userRepo.Update(ctx, resetToken.UserID.String(), map[string]interface{}{"password": string(hashedPassword)}); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}

	resetToken.Used = true
	if err := s.resetTokenRepo.UpdatePasswordResetToken(ctx, resetToken); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update token status", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset successfully", nil, nil)
}

func (s *service) OAuthLogin(ctx context.Context, provider string, gothUser goth.User) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	existingUser, err := s.userRepo.FindOne(ctx, map[string]interface{}{"email": gothUser.Email})

	var currentUser *user.User

	if err != nil {
		username := gothUser.NickName
		if username == "" {
			username = gothUser.Email
		}

		newUser := &user.User{
			ID:            uuid.New(),
			Username:      username,
			Email:         gothUser.Email,
			Password:      "",
			Status:        true,
			DefaultRoleID: 1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		newRoleUser := &user.UserRole{
			UserID: newUser.ID,
			RoleID: 1,
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
		}

		if err := s.userRepo.CreateUserRole(ctx, newRoleUser); err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user role", nil, nil)
		}

		// Reload the new user from DB to get populated Roles association
		loadedUser, err := s.userRepo.FindOne(ctx, map[string]interface{}{"id": newUser.ID})
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to load user data", nil, nil)
		}
		currentUser = loadedUser
	} else {
		if !existingUser.Status {
			return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
		}
		currentUser = existingUser
	}

	if len(currentUser.Roles) == 0 {
		return pkg.NewResponse(http.StatusForbidden, "User has no assigned roles", nil, nil)
	}

	userRoles := make([]string, len(currentUser.Roles))
	for i, role := range currentUser.Roles {
		userRoles[i] = role.Role
	}

	ttl := config.GetJWTTTL()
	claims := &jwt_pkg.UserJWTClaims{
		UserID:     currentUser.ID.String(),
		Role:       userRoles,
		ActiveRole: currentUser.DefaultRole.Role,
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

func (s *service) SwitchRole(ctx context.Context, claims jwt_pkg.UserJWTClaims, req SwitchRoleRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Role == "" {
		errValidation["role"] = "Role is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	// Validate the requested role is actually assigned to this user
	hasRole := false
	for _, r := range claims.Role {
		if r == req.Role {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return pkg.NewResponse(http.StatusForbidden, "Access denied: role not assigned to user", nil, nil)
	}

	claims.ActiveRole = req.Role
	newToken, err := jwt_pkg.GenerateJWTToken(claims)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate new token", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Role switched successfully", nil, AuthResponse{
		Token: newToken,
	})
}
