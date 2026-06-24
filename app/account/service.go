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
	GetAccountList(ctx context.Context, params AccountQueryParam, excludeSuperadmin bool) pkg.Response
	GetAccountByID(ctx context.Context, accountID string) pkg.Response
	SetAccountBanStatus(ctx context.Context, accountID string, payload SetAccountBanStatusRequest) pkg.Response

	AddAccountRole(ctx context.Context, accountID string, roleID int) pkg.Response
	UpdateAccountRole(ctx context.Context, accountID string, roleID int, payload UpdateAccountRoleRequest) pkg.Response

	UpdateUserProfile(ctx context.Context, accountID string, payload UpdateUserProfileRequest) pkg.Response
	UpdatePassword(ctx context.Context, accountID string, payload UpdatePasswordRequest) pkg.Response

	GetRoleList(ctx context.Context) pkg.Response
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

func (s *service) GetAccountList(ctx context.Context, params AccountQueryParam, excludeSuperadmin bool) pkg.Response {
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
		"exclude_superadmin": excludeSuperadmin,
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
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
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

	return pkg.NewResponse(http.StatusOK, "Daftar akun berhasil diambil", nil, toAccountListResponse(accounts, pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}))
}

func (s *service) GetAccountByID(ctx context.Context, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Format ID akun tidak valid", nil, nil)
	}

	account, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Detail akun berhasil diambil", nil, account.toUserProfileResponse())
}

func (s *service) SetAccountBanStatus(ctx context.Context, accountID string, payload SetAccountBanStatusRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Format ID akun tidak valid", nil, nil)
	}

	_, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	err = s.repo.UpdateAccount(ctx, accountID, map[string]interface{}{"is_banned": payload.BanStatus})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to update account banned status")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	if payload.BanStatus {
		return pkg.NewResponse(http.StatusOK, "Akun berhasil diblokir", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Blokir akun berhasil dilepas", nil, nil)
}

func (s *service) AddAccountRole(ctx context.Context, accountID string, roleID int) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Format ID akun tidak valid", nil, nil)
	}

	_, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to retrieve account")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	role, err := s.repo.FindOneRole(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Peran tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
			"role_id":   roleID,
		}).WithError(err).Error("failed to retrieve role")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	if role.ID == ProtectedSuperAdminRoleID {
		return pkg.NewResponse(http.StatusNotFound, "Peran tidak ditemukan", nil, nil)
	}

	_, err = s.repo.FindOneAccountRole(ctx, accountID, roleID)
	if err == nil {
		return pkg.NewResponse(http.StatusConflict, "Akun sudah memiliki peran ini", nil, nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.WithFields(logrus.Fields{"component": "account.service", "account_id": accountID, "role_id": roleID}).WithError(err).Error("failed to check existing account role")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
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
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Peran berhasil ditambahkan ke akun", nil, nil)
}

func (s *service) UpdateAccountRole(ctx context.Context, accountID string, roleID int, payload UpdateAccountRoleRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	role, err := s.repo.FindOneAccountRole(ctx, accountID, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Peran akun tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
			"role_id":    roleID,
		}).WithError(err).Error("failed to retrieve account role")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	if payload.IsActive == role.IsActive {
		msg := "Peran akun sudah dinonaktifkan"
		if payload.IsActive {
			msg = "Peran akun sudah diaktifkan"
		}
		return pkg.NewResponse(http.StatusOK, msg, nil, nil)
	}

	if !payload.IsActive {
		if role.IsDefault {
			return pkg.NewResponse(http.StatusForbidden, "Tidak dapat menonaktifkan peran utama", nil, nil)
		}

		count, err := s.repo.CountActiveAccountRoles(ctx, accountID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "account.service",
				"account_id": accountID,
			}).WithError(err).Error("failed to retrieve account roles")
			return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
		}

		if count <= 1 {
			return pkg.NewResponse(http.StatusForbidden, "Akun harus memiliki setidaknya satu peran aktif", nil, nil)
		}
	}

	if err := s.repo.UpdateAccountRole(ctx, accountID, roleID, map[string]interface{}{"is_active": payload.IsActive}); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
			"role_id":    roleID,
		}).WithError(err).Error("failed to update account role status")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	if payload.IsActive {
		return pkg.NewResponse(http.StatusOK, "Peran akun berhasil diaktifkan", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Peran akun berhasil dinonaktifkan", nil, nil)
}

func (s *service) UpdateUserProfile(ctx context.Context, accountID string, payload UpdateUserProfileRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	errValidation := make(map[string]string)
	updateProfileMap := make(map[string]interface{})
	updateAccountMap := make(map[string]interface{})

	if payload.Username != "" {
		if len(payload.Username) > 20 {
			errValidation["username"] = "Nama pengguna maksimal 20 karakter"
		} else if len(payload.Username) < 3 {
			errValidation["username"] = "Nama pengguna minimal 3 karakter"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"username": payload.Username})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["username"] = "Nama pengguna sudah digunakan"
			} else {
				updateProfileMap["username"] = payload.Username
			}
		}
	}
	if payload.Email != "" {
		if !pkg.IsValidEmail(payload.Email) {
			errValidation["email"] = "Format email tidak valid"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"email": payload.Email})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["email"] = "Email sudah digunakan"
			} else {
				updateAccountMap["email"] = payload.Email
			}
		}
	}
	if payload.Phone != "" {
		if !pkg.IsValidPhoneNumber(payload.Phone) {
			errValidation["phone"] = "Format nomor telepon tidak valid"
		} else {
			existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"phone": payload.Phone})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
			}
			if err == nil && existing.ID.String() != accountID {
				errValidation["phone"] = "Nomor telepon sudah digunakan"
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
				errValidation["defaultAccountRoleId"] = "ID peran utama tidak valid"
			} else {
				logrus.WithFields(logrus.Fields{
					"component":  "account.service",
					"account_id": accountID,
					"role_id":    payload.DefaultAccountRoleID,
				}).WithError(err).Error("failed to retrieve account role")
				return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil peran akun", nil, nil)
			}
		} else if !role.IsActive {
			errValidation["defaultAccountRoleId"] = "Tidak dapat menetapkan peran tidak aktif sebagai peran utama"
		} else if payload.DefaultAccountRoleID != ProtectedSuperAdminRoleID {
			_, err := s.repo.FindOneAccountRole(ctx, accountID, ProtectedSuperAdminRoleID)
			if err == nil {
				errValidation["defaultAccountRoleId"] = "Superadmin tidak dapat mengubah peran utama mereka"
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				logrus.WithFields(logrus.Fields{
					"component":  "account.service",
					"account_id": accountID,
				}).WithError(err).Error("failed to check superadmin status")
				return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
			}
		}
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	var uploadedURL string
	var oldProfilePicture string

	if payload.ProfilePicture != nil {
		existing, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
		if err != nil {
			return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
		}

		oldProfilePicture = s3_pkg.ExtractObjectNameFromURL(existing.UserProfile.ProfilePicture)

		uploadedURL, err = s.s3Client.UploadFile(ctx, payload.ProfilePicture, "accounts")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "account.service",
				"account_id": accountID,
			}).WithError(err).Error("failed to upload profile picture")
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah gambar profil", nil, nil)
		}
		updateProfileMap["profile_picture"] = uploadedURL
	}

	if err := s.repo.UpdateFullProfile(ctx, accountID, updateAccountMap, updateProfileMap, payload.DefaultAccountRoleID); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Error("failed to update full profile")

		// Clean up uploaded S3 file since DB update failed
		if uploadedURL != "" {
			if cleanupErr := s.s3Client.DeleteFile(ctx, uploadedURL); cleanupErr != nil {
				logrus.WithError(cleanupErr).Error("failed to clean up uploaded S3 image after DB failure")
			}
		}

		if err.Error() == "account not found" {
			return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	if uploadedURL != "" && oldProfilePicture != "" {
		if deleteErr := s.s3Client.DeleteFile(ctx, oldProfilePicture); deleteErr != nil {
			logrus.WithError(deleteErr).Warnf("failed to delete orphaned S3 image: %s", oldProfilePicture)
		}
	}

	return pkg.NewResponse(http.StatusOK, "Profil berhasil diperbarui", nil, nil)
}

func (s *service) UpdatePassword(ctx context.Context, accountID string, payload UpdatePasswordRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	account, err := s.repo.FindOneAccount(ctx, map[string]interface{}{"id": accountID})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": accountID,
		}).WithError(err).Warn("account not found for password update")
		return pkg.NewResponse(http.StatusNotFound, "Akun tidak ditemukan", nil, nil)
	}

	errValidation := make(map[string]string)
	hasExistingPassword := account.Password != ""

	if hasExistingPassword {
		if payload.CurrentPassword == "" {
			errValidation["currentPassword"] = "Kata sandi saat ini wajib diisi"
		}
	}

	if payload.NewPassword == "" {
		errValidation["newPassword"] = "Kata sandi baru wajib diisi"
	} else if hasExistingPassword && payload.CurrentPassword == payload.NewPassword {
		errValidation["newPassword"] = "Kata sandi baru tidak boleh sama dengan kata sandi saat ini"
	} else if !pkg.IsValidLengthPassword(payload.NewPassword) {
		errValidation["newPassword"] = "Kata sandi baru minimal 8 karakter"
	} else if !pkg.IsStrongPassword(payload.NewPassword) {
		errValidation["newPassword"] = "Kata sandi baru harus mengandung setidaknya satu huruf besar, satu huruf kecil, satu angka, dan satu karakter khusus"
	}

	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	if hasExistingPassword {
		if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(payload.CurrentPassword)); err != nil {
			errValidation["currentPassword"] = "Kata sandi saat ini salah"
			return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
		}).WithError(err).Error("failed to hash password")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memproses kata sandi baru", nil, nil)
	}

	updateMap := map[string]interface{}{
		"password": string(hashedPassword),
	}
	if !hasExistingPassword {
		updateMap["email_verified"] = true
	}

	err = s.repo.UpdateAccount(ctx, account.ID.String(), updateMap)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "account.service",
			"account_id": account.ID.String(),
		}).WithError(err).Error("failed to update password in database")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui kata sandi", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Kata sandi berhasil diperbarui", nil, nil)
}

func (s *service) GetRoleList(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	roles, err := s.repo.FindAllRoles(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "account.service",
		}).WithError(err).Error("failed to get all roles")
		return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Daftar peran berhasil diambil", nil, toRolesResponse(roles))
}
