package account

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	GetAccountList(ctx context.Context, params AccountQueryParam) pkg.Response
	GetAccountByID(ctx context.Context, accountID string) pkg.Response
	SetAccountBanStatus(ctx context.Context, accountID string, payload SetAccountBanStatusRequest) pkg.Response

	AddAccountRole(ctx context.Context, accountID string, roleID int) pkg.Response
	UpdateAccountRole(ctx context.Context, accountID string, roleID int, payload UpdateAccountRoleRequest) pkg.Response

	UpdateUserProfile(ctx context.Context, accountID string, payload UpdateUserProfileRequest) pkg.Response
	UpdatePassword(ctx context.Context, accountID string, payload UpdatePasswordRequest) pkg.Response
}

type service struct {
	repo     Repository
	timeout  time.Duration
	s3Client s3_pkg.Client
}

func NewService(r Repository, timeout time.Duration, s3Client s3_pkg.Client) Service {
	return &service{
		repo:     r,
		timeout:  timeout,
		s3Client: s3Client,
	}
}

func (s *service) GetAccountList(ctx context.Context, params AccountQueryParam) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	options := map[string]interface{}{
		"limit":              params.Limit,
		"exclude_superadmin": true,
	}
	if params.RoleID != 0 {
		options["role_id"] = params.RoleID
	}
	if params.IsBanned != nil {
		options["is_banned"] = *params.IsBanned
	}
	if params.Search != "" {
		options["search"] = params.Search
	}
	if params.SortOrder != "" {
		options["sort_order"] = params.SortOrder
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	accounts, err := s.repo.FindAllAccounts(ctx, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
			"params":    params,
		}).WithError(err).Error("failed to retrieve users")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	hasMore := len(accounts) > params.Limit
	if hasMore {
		accounts = accounts[:params.Limit]
	}
	if params.PrevCursor != "" {
		for i, j := 0, len(accounts)-1; i < j; i, j = i+1, j-1 {
			accounts[i], accounts[j] = accounts[j], accounts[i]
		}
	}

	usingPrevCursor := params.PrevCursor != ""
	var nextCursor, prevCursor string
	hasNext := false
	hasPrev := false

	if usingPrevCursor {
		hasPrev = hasMore
		hasNext = true
	} else {
		hasNext = hasMore
		hasPrev = params.NextCursor != ""
	}

	if len(accounts) > 0 {
		first := accounts[0]
		last := accounts[len(accounts)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	return pkg.NewResponse(http.StatusOK, "Accounts list retrieved successfully", nil, toAccountListResponse(accounts, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetAccountByID(ctx context.Context, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid account ID format", nil, nil)
	}

	account, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Account details retrieved successfully", nil, account.toUserProfileResponse())
}

func (s *service) SetAccountBanStatus(ctx context.Context, accountID string, payload SetAccountBanStatusRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid account ID format", nil, nil)
	}

	_, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	err = s.repo.UpdateAccount(ctx, accountID, map[string]interface{}{"is_banned": payload.BanStatus})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to update account banned status")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if payload.BanStatus {
		return pkg.NewResponse(http.StatusOK, "Account banned successfully", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Account unbanned successfully", nil, nil)
}

func (s *service) AddAccountRole(ctx context.Context, accountID string, roleID int) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid account ID format", nil, nil)
	}

	_, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	role, err := s.repo.FindOneRole(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Role not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
			"role_id":   roleID,
		}).WithError(err).Error("failed to retrieve role")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if role.ID == ProtectedSuperAdminRoleID {
		return pkg.NewResponse(http.StatusNotFound, "Role not found", nil, nil)
	}

	_, err = s.repo.FindOneAccountRole(ctx, accountID, roleID)
	if err == nil {
		return pkg.NewResponse(http.StatusConflict, "Account already has this role", nil, nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.WithFields(logrus.Fields{"component": "account.service", "account_id": accountID, "role_id": roleID}).WithError(err).Error("failed to check existing account role")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	newRole := &AccountRole{
		AccountID: uuid.MustParse(accountID),
		RoleID:    roleID,
	}

	if err := s.repo.CreateAccountRole(ctx, newRole); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
			"role_id":    roleID,
		}).WithError(err).Error("failed to create account role")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Role added to account successfully", nil, nil)
}

func (s *service) UpdateAccountRole(ctx context.Context, accountID string, roleID int, payload UpdateAccountRoleRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	role, err := s.repo.FindOneAccountRole(ctx, accountID, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Account role not found", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
			"role_id":    roleID,
		}).WithError(err).Error("failed to retrieve account role")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if payload.IsActive == role.IsActive {
		msg := "Account role is already deactivated"
		if payload.IsActive {
			msg = "Account role is already activated"
		}
		return pkg.NewResponse(http.StatusOK, msg, nil, nil)
	}

	if !payload.IsActive {
		if role.IsDefault {
			return pkg.NewResponse(http.StatusForbidden, "Cannot deactivate default role", nil, nil)
		}

		count, err := s.repo.CountActiveAccountRoles(ctx, accountID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "account.service",
				"account_id": accountID,
			}).WithError(err).Error("failed to retrieve account roles")
			return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
		}

		if count <= 1 {
			return pkg.NewResponse(http.StatusForbidden, "Account must have at least one active role", nil, nil)
		}
	}

	if err := s.repo.UpdateAccountRole(ctx, accountID, roleID, map[string]interface{}{"is_active": payload.IsActive}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
			"role_id":    roleID,
		}).WithError(err).Error("failed to update account role status")
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if payload.IsActive {
		return pkg.NewResponse(http.StatusOK, "Account role activated successfully", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Account role deactivated successfully", nil, nil)
}

func (s *service) UpdateUserProfile(ctx context.Context, accountID string, payload UpdateUserProfileRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	updateProfileMap := make(map[string]interface{})
	updateAccountMap := make(map[string]interface{})

	if payload.Username != "" {
		if len(payload.Username) > 20 {
			errValidation["username"] = "Username must be at most 20 characters"
		} else if len(payload.Username) < 3 {
			errValidation["username"] = "Username must be at least 3 characters"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"username": payload.Username})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["username"] = "Username is already taken"
			} else {
				updateProfileMap["username"] = payload.Username
			}
		}
	}
	if payload.Email != "" {
		if !pkg.IsValidEmail(payload.Email) {
			errValidation["email"] = "Invalid email format"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"email": payload.Email})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["email"] = "Email is already taken"
			} else {
				updateAccountMap["email"] = payload.Email
			}
		}
	}
	if payload.Phone != "" {
		if !pkg.IsValidPhoneNumber(payload.Phone) {
			errValidation["phone"] = "Invalid phone number format"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"phone": payload.Phone})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["phone"] = "Phone number is already taken"
			} else {
				updateProfileMap["phone"] = payload.Phone
			}
		}
	}
	if payload.Address != "" {
		updateProfileMap["address"] = payload.Address
	}
	if payload.DefaultAccountRoleID != 0 {
		role, err := s.repo.FindOneAccountRole(ctx, accountID, payload.DefaultAccountRoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				errValidation["default_account_role_id"] = "Invalid default account role id"
			} else {
				logrus.WithFields(logrus.Fields{
					"component":  "account.service",
					"account_id": accountID,
					"role_id":    payload.DefaultAccountRoleID,
				}).WithError(err).Error("failed to retrieve account role")
				return pkg.NewResponse(http.StatusInternalServerError, "Failed to retrieve account role", nil, nil)
			}
		} else if !role.IsActive {
			errValidation["default_account_role_id"] = "Cannot set an inactive role as default"
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	var oldProfilePicture string
	var newProfilePicture string
	if payload.ProfilePicture != nil {
		existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
		if err != nil {
			return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
		}

		oldProfilePicture = s3_pkg.ExtractObjectNameFromURL(existing.UserProfile.ProfilePicture)

		uploadedURL, err := s.s3Client.UploadFile(ctx, payload.ProfilePicture, "accounts")
		if err != nil {
			return pkg.NewResponse(http.StatusInternalServerError, "Failed to upload image", nil, nil)
		}
		updateProfileMap["profile_picture"] = uploadedURL
		newProfilePicture = s3_pkg.ExtractObjectNameFromURL(uploadedURL)
	}

	if err := s.repo.UpdateFullProfile(ctx, accountID, updateAccountMap, updateProfileMap, payload.DefaultAccountRoleID); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to update full profile")

		if newProfilePicture != "" {
			go func(key string) {
				if err := s.s3Client.DeleteFile(context.Background(), key); err != nil {
					logrus.WithError(err).Error("failed to clean up newly uploaded s3 image after db failure")
				}
			}(newProfilePicture)
		}

		if err.Error() == "account not found" {
			return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Internal server error", nil, nil)
	}

	if payload.ProfilePicture != nil && oldProfilePicture != "" {
		go func(key string) {
			if err := s.s3Client.DeleteFile(context.Background(), key); err != nil {
				logrus.WithError(err).Warnf("failed to delete orphaned s3 image: %s", key)
			}
		}(oldProfilePicture)
	}

	return pkg.NewResponse(http.StatusOK, "Profile updated successfully", nil, nil)
}

func (s *service) UpdatePassword(ctx context.Context, accountID string, payload UpdatePasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	if payload.CurrentPassword == "" {
		errValidation["current_password"] = "Current password is required"
	}
	if payload.NewPassword == "" {
		errValidation["new_password"] = "New password is required"
	} else if payload.CurrentPassword == payload.NewPassword {
		errValidation["new_password"] = "New password cannot be the same as the current password"
	} else if !pkg.IsValidLengthPassword(payload.NewPassword) {
		errValidation["new_password"] = "New password must be at least 8 characters long"
	} else if !pkg.IsStrongPassword(payload.NewPassword) {
		errValidation["new_password"] = "New password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	account, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Warn("account not found for password update")
		return pkg.NewResponse(http.StatusNotFound, "Account not found", nil, nil)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(payload.CurrentPassword)); err != nil {
		errValidation["current_password"] = "Current password is incorrect"
		return pkg.NewResponse(http.StatusBadRequest, "Validation error", errValidation, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
		}).WithError(err).Error("failed to hash password")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash new password", nil, nil)
	}

	err = s.repo.UpdateAccount(ctx, account.ID.String(), map[string]interface{}{"password": string(hashedPassword)})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": account.ID.String(),
		}).WithError(err).Error("failed to update password in database")
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to update password", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Password updated successfully", nil, nil)
}
