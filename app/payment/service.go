package payment

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetPaymentMethods(ctx context.Context) pkg.Response
	UpdatePaymentMethod(ctx context.Context, id int, payload UpdatePaymentMethodRequest) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *service) GetPaymentMethods(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pms, err := s.repo.FindAll(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "payment.service",
		}).WithError(err).Error("failed to retrieve payment methods")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data metode pembayaran", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Sukses", nil, ToPaymentMethodListResponse(pms))
}

func (s *service) UpdatePaymentMethod(ctx context.Context, id int, payload UpdatePaymentMethodRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Check if payment method exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Metode pembayaran tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "payment.service",
			"id":        id,
		}).WithError(err).Error("failed to find payment method")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan metode pembayaran", nil, nil)
	}

	errValidation := make(map[string]string)
	switch payload.FeeType {
	case FeeTypeFlat:
		if payload.FeeValue < 0 || payload.FeeValue > 10000 {
			errValidation["feeValue"] = "Nilai biaya flat tidak boleh kurang dari 0 atau lebih dari 10.000"
		}
	case FeeTypePercentage:
		if payload.FeeValue < 0 || payload.FeeValue > 0.05 {
			errValidation["feeValue"] = "Nilai biaya persentase tidak boleh kurang dari 0% atau lebih dari 5% (0.05)"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", errValidation, nil)
	}

	updates := map[string]interface{}{
		"fee_type":  payload.FeeType,
		"fee_value": payload.FeeValue,
		"is_active": *payload.IsActive,
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "payment.service",
			"id":        id,
		}).WithError(err).Error("failed to update payment method")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui metode pembayaran", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Metode pembayaran berhasil diperbarui", nil, nil)
}
