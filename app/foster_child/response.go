package foster_child

import "github.com/Vilamuzz/yota-backend/pkg"

type FosterChildResponse struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	ImageURL  string   `json:"image_url"`
	Age       int      `json:"age"`
	Gender    string   `json:"gender"`
	Status    bool     `json:"status"` // true for not graduated, false for graduated
	Category  Category `json:"category"`
	BirthDate string   `json:"birth_date"`
	Address   string   `json:"address"`
}

type FosterChildListResponse struct {
	FosterChildren []FosterChildResponse `json:"foster_children"`
	Pagination     pkg.CursorPagination  `json:"pagination"`
}

func (f *FosterChild) ToFosterChildResponse() FosterChildResponse {
	return FosterChildResponse{
		ID:        f.ID,
		Name:      f.Name,
		ImageURL:  f.ImageURL,
		Age:       f.Age,
		Gender:    f.Gender,
		Status:    f.Status,
		Category:  f.Category,
		BirthDate: f.BirthDate.Format("2006-01-02"),
		Address:   f.Address,
	}
}

func ToFosterChildListResponse(fosterChildren []FosterChild, pagination pkg.CursorPagination) FosterChildListResponse {
	var responses []FosterChildResponse
	for _, f := range fosterChildren {
		responses = append(responses, f.ToFosterChildResponse())
	}
	if responses == nil {
		responses = []FosterChildResponse{}
	}
	return FosterChildListResponse{
		FosterChildren: responses,
		Pagination:     pagination,
	}
}
