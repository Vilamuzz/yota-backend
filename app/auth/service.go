package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Register(ctx context.Context, payload RegisterRequest) pkg.Response
	Login(ctx context.Context, payload LoginRequest) pkg.Response
	ForgetPassword(ctx context.Context, payload ForgetPasswordRequest) pkg.Response
	ResetPassword(ctx context.Context, payload ResetPasswordRequest) pkg.Response
	OAuthLogin(ctx context.Context, provider string, gothUser goth.User) pkg.Response
	VerifyEmail(ctx context.Context, token string) pkg.Response
	ResendVerificationEmail(ctx context.Context, email string) pkg.Response
	SwitchRole(ctx context.Context, claims jwt_pkg.UserJWTClaims, payload SwitchRoleRequest) pkg.Response
}

type service struct {
	accountRepo    account.Repository
	authRepo       Repository
	emailService   *pkg.EmailService
	contextTimeout time.Duration
}

func NewService(authRepo Repository, accountRepo account.Repository, timeout time.Duration) Service {
	return &service{
		accountRepo:    accountRepo,
		authRepo:       authRepo,
		emailService:   pkg.NewEmailService(),
		contextTimeout: timeout,
	}
}

func (s *service) Register(ctx context.Context, payload RegisterRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	payload.Username = strings.TrimSpace(payload.Username)
	payload.Email = strings.TrimSpace(payload.Email)

	if len(payload.Username) > 20 {
		errValidation["username"] = "Username must be at most 20 characters"
	} else if len(payload.Username) < 3 {
		errValidation["username"] = "Username must be at least 3 characters"
	} else {
		_, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"username": payload.Username})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
		}
		if err == nil {
			errValidation["username"] = "Username is already taken"
		}
	}

	if payload.Email == "" {
		errValidation["email"] = "Email is required"
	} else if !pkg.IsValidEmail(payload.Email) {
		errValidation["email"] = "Invalid email format"
	} else {
		existingUser, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"email": payload.Email})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     payload.Email,
			}).WithError(err).Error("failed to retrieve account by email")
			return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
		}
		if err == nil && existingUser != nil {
			errValidation["email"] = "Email is already registered"
		}
	}

	if payload.Password == "" {
		errValidation["password"] = "Password is required"
	} else if !pkg.IsValidLengthPassword(payload.Password) {
		errValidation["password"] = "Password must be at least 8 characters"
	} else if !pkg.IsStrongPassword(payload.Password) {
		errValidation["password"] = "Password must contain uppercase, lowercase, and number"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
		}).WithError(err).Error("failed to hash password")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	now := time.Now()
	accountID := uuid.New()

	newAccount := &account.Account{
		ID:            accountID,
		Email:         payload.Email,
		Password:      string(hashedPassword),
		IsBanned:      false,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		UserProfile: account.UserProfile{
			ID:        uuid.New(),
			Username:  payload.Username,
			CreatedAt: now,
			UpdatedAt: now,
		},
		AccountRoles: []account.AccountRole{
			{
				RoleID:    account.OrangTuaAsuhRoleID,
				IsDefault: true,
				IsActive:  true,
			},
		},
	}

	if err := s.accountRepo.CreateAccount(ctx, newAccount); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
			"email":     payload.Email,
		}).WithError(err).Error("failed to create user account")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
		}).WithError(err).Error("failed to generate verification token")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate verification token", nil, nil)
	}

	verificationToken := hex.EncodeToString(tokenBytes)
	emailVerification := &EmailVerificationToken{
		ID:        uuid.New(),
		AccountID: newAccount.ID,
		Token:     verificationToken,
		ExpiredAt: time.Now().Add(24 * time.Hour),
		IsUsed:    false,
	}

	if err := s.authRepo.CreateEmailVerificationToken(ctx, emailVerification); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
			"email":     payload.Email,
		}).WithError(err).Error("failed to create email verification token")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create verification token", nil, nil)
	}

	go func(email, username, token string) {
		if err := s.emailService.SendEmailVerification(email, username, token); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     email,
			}).WithError(err).Error("failed to send verification email asynchronously")
		}
	}(newAccount.Email, newAccount.UserProfile.Username, verificationToken)

	return pkg.NewResponse(http.StatusCreated, "Registration successful. Please check your email to verify your account.", nil, map[string]interface{}{
		"email": newAccount.Email,
	})
}

func (s *service) Login(ctx context.Context, payload LoginRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	payload.Email = strings.TrimSpace(payload.Email)

	if payload.Email == "" {
		errValidation["email"] = "Email is required"
	} else if !pkg.IsValidEmail(payload.Email) {
		errValidation["email"] = "Invalid email format"
	}

	if payload.Password == "" {
		errValidation["password"] = "Password is required"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	existingUser, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"email": payload.Email})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
			"email":     payload.Email,
		}).WithError(err).Error("failed to retrieve account by email during login")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(payload.Password)); err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Invalid email or password", nil, nil)
	}

	if !existingUser.EmailVerified {
		return pkg.NewResponse(http.StatusForbidden, "Please verify your email before logging in", nil, nil)
	}

	if existingUser.IsBanned {
		return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
	}

	ttl := config.GetJWTTTL()
	var userRoles []enum.RoleName
	var activeRole enum.RoleName

	if len(existingUser.AccountRoles) > 0 {
		for _, role := range existingUser.AccountRoles {
			if role.IsActive {
				userRoles = append(userRoles, role.Role.Name)
				if role.IsDefault {
					activeRole = role.Role.Name
				}
			}
		}
		if activeRole == "" && len(userRoles) > 0 {
			activeRole = userRoles[0]
		}
	}

	if len(userRoles) == 0 {
		return pkg.NewResponse(http.StatusForbidden, "Your account has no active roles", nil, nil)
	}

	claims := &jwt_pkg.UserJWTClaims{
		AccountID:  existingUser.ID.String(),
		Roles:      userRoles,
		ActiveRole: activeRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt_pkg.GenerateJWTToken(claims)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": existingUser.ID.String(),
		}).WithError(err).Error("failed to generate jwt token")
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

	verificationToken, err := s.authRepo.FetchEmailVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"token": "Invalid or expired verification token"}, nil)
		} else {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
			}).WithError(err).Error("failed to fetch email verification token")
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch email verification token", nil, nil)
		}
	}

	errValidation := make(map[string]string)
	if verificationToken.IsUsed {
		errValidation["token"] = "Verification token has already been used"
	} else if time.Now().After(verificationToken.ExpiredAt) {
		errValidation["token"] = "Verification token has expired"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	if err := s.accountRepo.UpdateAccount(ctx, verificationToken.AccountID.String(), map[string]interface{}{"email_verified": true}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": verificationToken.AccountID.String(),
		}).WithError(err).Error("failed to update account email_verified status")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to verify email", nil, nil)
	}

	verificationToken.IsUsed = true
	if err := s.authRepo.UpdateEmailVerificationToken(ctx, verificationToken); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": verificationToken.AccountID.String(),
			"token":      token,
		}).WithError(err).Error("failed to mark verification token as used")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update token status", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Email verified successfully", nil, nil)
}

func (s *service) ResendVerificationEmail(ctx context.Context, email string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	existingUser, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"email": email})
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
		AccountID: existingUser.ID,
		Token:     verificationToken,
		ExpiredAt: time.Now().Add(24 * time.Hour),
		IsUsed:    false,
	}

	if err := s.authRepo.CreateEmailVerificationToken(ctx, emailVerification); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create verification token", nil, nil)
	}

	go func(email, username, token string) {
		if err := s.emailService.SendEmailVerification(email, username, token); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     email,
			}).WithError(err).Error("failed to send verification email asynchronously during resend")
		}
	}(existingUser.Email, existingUser.UserProfile.Username, verificationToken)

	return pkg.NewResponse(http.StatusOK, "Verification email sent successfully", nil, nil)
}

func (s *service) ForgetPassword(ctx context.Context, payload ForgetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	payload.Email = strings.TrimSpace(payload.Email)
	if payload.Email == "" || !pkg.IsValidEmail(payload.Email) {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", map[string]string{"email": "Invalid email format"}, nil)
	}
	const successMsg = "If the email exists, a password reset link has been sent to your inbox."

	existingUser, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"email": payload.Email})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusOK, successMsg, nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
		}).WithError(err).Error("failed to find account by email during forget password")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logrus.WithError(err).Error("failed to generate reset token")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate reset token", nil, nil)
	}

	resetToken := hex.EncodeToString(tokenBytes)
	passwordReset := &PasswordResetToken{
		ID:        uuid.New(),
		AccountID: existingUser.ID,
		Token:     resetToken,
		ExpiredAt: time.Now().Add(1 * time.Hour),
		IsUsed:    false,
	}

	if err := s.authRepo.CreatePasswordResetToken(ctx, passwordReset); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": existingUser.ID.String(),
		}).WithError(err).Error("failed to create reset token")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create reset token", nil, nil)
	}

	go func(email, username, token string) {
		if err := s.emailService.SendPasswordResetEmail(email, username, token); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     email,
			}).WithError(err).Error("failed to send reset email")
		}
	}(payload.Email, existingUser.UserProfile.Username, resetToken)

	return pkg.NewResponse(http.StatusOK, successMsg, nil, nil)
}

func (s *service) ResetPassword(ctx context.Context, payload ResetPasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.NewPassword == "" {
		errValidation["new_password"] = "New password is required"
	} else if !pkg.IsValidLengthPassword(payload.NewPassword) {
		errValidation["new_password"] = "Password must be at least 8 characters"
	} else if !pkg.IsStrongPassword(payload.NewPassword) {
		errValidation["new_password"] = "Password must contain uppercase, lowercase, and number"
	}

	resetToken, err := s.authRepo.FetchPasswordResetToken(ctx, payload.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errValidation["token"] = "Invalid or expired reset token"
		} else {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
			}).WithError(err).Error("failed to fetch reset token")
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch reset token", nil, nil)
		}
	} else {
		if resetToken.IsUsed {
			errValidation["token"] = "Reset token already used"
		} else if time.Now().After(resetToken.ExpiredAt) {
			errValidation["token"] = "Reset token has expired"
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "auth.service",
		}).WithError(err).Error("failed to hash password")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	if err := s.authRepo.ResetPassword(ctx, resetToken.AccountID.String(), resetToken.ID.String(), string(hashedPassword)); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": resetToken.AccountID.String(),
			"token":      payload.Token,
		}).WithError(err).Error("failed to execute password reset transaction")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to reset password", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password reset successfully", nil, nil)
}

func (s *service) OAuthLogin(ctx context.Context, provider string, gothUser goth.User) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	existingUser, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"email": gothUser.Email})
	var currentAccount *account.Account

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     gothUser.Email,
				"provider":  provider,
			}).WithError(err).Error("failed to find account during oauth login")
			return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
		}

		username := gothUser.NickName
		if username == "" {
			parts := strings.Split(gothUser.Email, "@")
			username = parts[0]
		}
		existingAccount, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"user_profile.username": username})
		if err == nil && existingAccount != nil {
			username = fmt.Sprintf("%s_%s", username, strings.Replace(uuid.New().String(), "-", "", -1)[:6])
		}

		accountID := uuid.New()
		now := time.Now()
		newAccount := &account.Account{
			ID:            accountID,
			Email:         gothUser.Email,
			Password:      "",
			IsBanned:      false,
			EmailVerified: true,
			CreatedAt:     now,
			UpdatedAt:     now,
			UserProfile: account.UserProfile{
				ID:        uuid.New(),
				Username:  username,
				CreatedAt: now,
				UpdatedAt: now,
			},
			AccountRoles: []account.AccountRole{
				{
					RoleID:    account.OrangTuaAsuhRoleID,
					IsDefault: true,
					IsActive:  true,
				},
			},
		}

		if err := s.accountRepo.CreateAccount(ctx, newAccount); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "auth.service",
				"email":     gothUser.Email,
				"provider":  provider,
			}).WithError(err).Error("failed to create oauth user")
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to create user", nil, nil)
		}

		currentAccount = newAccount
	} else {
		if existingUser.IsBanned {
			return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
		}
		currentAccount = existingUser
	}

	ttl := config.GetJWTTTL()
	var userRoles []enum.RoleName
	var activeRole enum.RoleName

	if len(currentAccount.AccountRoles) > 0 {
		for _, role := range currentAccount.AccountRoles {
			if role.IsActive {
				userRoles = append(userRoles, role.Role.Name)
				if role.IsDefault {
					activeRole = role.Role.Name
				}
			}
		}
	}

	if len(userRoles) == 0 {
		return pkg.NewResponse(http.StatusForbidden, "Your account has no active roles", nil, nil)
	}

	claims := &jwt_pkg.UserJWTClaims{
		AccountID:  currentAccount.ID.String(),
		Roles:      userRoles,
		ActiveRole: activeRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt_pkg.GenerateJWTToken(claims)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "auth.service",
			"account_id": currentAccount.ID.String(),
		}).WithError(err).Error("failed to generate oauth jwt token")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate token", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "OAuth login successful", nil, AuthResponse{Token: token})
}

func (s *service) SwitchRole(ctx context.Context, claims jwt_pkg.UserJWTClaims, payload SwitchRoleRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	accountID := claims.AccountID
	account, err := s.accountRepo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusUnauthorized, "Account no longer exists", nil, nil)
		}
		logrus.WithFields(logrus.Fields{"component": "auth.service", "account_id": accountID}).WithError(err).Error("failed to retrieve account during role switch")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if account.IsBanned {
		return pkg.NewResponse(http.StatusForbidden, "Your account has been banned", nil, nil)
	}

	hasRole := false
	var activeUserRoles []enum.RoleName

	for _, role := range account.AccountRoles {
		if role.IsActive {
			activeUserRoles = append(activeUserRoles, role.Role.Name)
			if string(role.Role.Name) == payload.Role {
				hasRole = true
			}
		}
	}
	if !hasRole {
		return pkg.NewResponse(http.StatusForbidden, "Access denied: role not assigned to user", nil, nil)
	}

	ttl := config.GetJWTTTL()
	newClaims := &jwt_pkg.UserJWTClaims{
		AccountID:  account.ID.String(),
		Roles:      activeUserRoles,
		ActiveRole: enum.RoleName(payload.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	newToken, err := jwt_pkg.GenerateJWTToken(newClaims)
	if err != nil {
		logrus.WithFields(logrus.Fields{"component": "auth.service", "account_id": accountID}).WithError(err).Error("failed to generate new token during role switch")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to generate new token", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Role switched successfully", nil, AuthResponse{
		Token: newToken,
	})
}
