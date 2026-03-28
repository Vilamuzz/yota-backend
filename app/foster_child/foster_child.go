package foster_child

import "time"

type FosterChild struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	Status    bool      `json:"status"` // true for not graduated, false for graduated
	Category  Category  `json:"category"`
	BirthDate time.Time `json:"birth_date"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category string

const (
	CategoryFatherless Category = "yatim"
	CategoryMotherless Category = "piatu"
	CategoryOrphan     Category = "yatim piatu"
)
