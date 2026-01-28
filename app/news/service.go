package news

type Service interface {
	FetchAllNews() ([]News, error)
	Create(news *News) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) FetchAllNews() ([]News, error) {
	return s.repo.FetchAllNews()
}

func (s *service) Create(news *News) error {
	return s.repo.Create(news)
}
