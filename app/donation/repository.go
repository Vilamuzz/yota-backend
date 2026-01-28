package donation

import "gorm.io/gorm"

type Repository interface {
	FetchAllDonations() ([]Donation, error)
	Create(donation *Donation) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FetchAllDonations() ([]Donation, error) {
	var donations []Donation
	if err := r.Conn.Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *repository) Create(donation *Donation) error {
	return r.Conn.Create(donation).Error
}
