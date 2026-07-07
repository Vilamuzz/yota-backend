package models

import (
	"fmt"

	"github.com/Vilamuzz/yota-backend/app/payment"
	"gorm.io/gorm"
)

func SeedPaymentMethods(db *gorm.DB) error {
	fmt.Println("Seeding payment methods...")
	paymentMethods := []payment.PaymentMethods{
		{Code: "bca_va", Name: "Bank BCA", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "bri_va", Name: "Bank BRI", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "bni_va", Name: "Bank BNI", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "mandiri_va", Name: "Bank Mandiri", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "permata_va", Name: "Bank Permata", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "cimb_va", Name: "Bank CIMB", FeeType: payment.FeeTypeFlat, FeeValue: 4000, IsActive: true},
		{Code: "qris", Name: "QRIS", FeeType: payment.FeeTypePercentage, FeeValue: 0.007, IsActive: true},
		{Code: "gopay", Name: "GoPay", FeeType: payment.FeeTypePercentage, FeeValue: 0.02, IsActive: true},
		{Code: "shopeepay", Name: "ShopeePay", FeeType: payment.FeeTypePercentage, FeeValue: 0.02, IsActive: true},
		{Code: "dana", Name: "DANA", FeeType: payment.FeeTypePercentage, FeeValue: 0.015, IsActive: true},
		{Code: "ovo", Name: "OVO", FeeType: payment.FeeTypePercentage, FeeValue: 0.015, IsActive: true},
	}

	for _, paymentMethod := range paymentMethods {
		if err := db.FirstOrCreate(&paymentMethod, payment.PaymentMethods{Code: paymentMethod.Code}).Error; err != nil {
			return fmt.Errorf("failed to seed payment method '%s': %w", paymentMethod.Name, err)
		}
	}
	return nil
}
