package service

import "github.com/ingoxx/stock-backend/internal/domain"

type GoldenService struct {
	repo domain.GoldenRepository
}

func NewUserService(repo domain.GoldenRepository) *GoldenService {
	return &GoldenService{repo: repo}
}

func (gs *GoldenService) GetGoldenInfo() ([]*domain.GoldenInfo, error) {
	return gs.GetGoldenInfo()
}

func (gs *GoldenService) SetGoldenDiffPrice(price float64) error {
	return gs.repo.SetGoldenBuyPrice(price)
}

func (gs *GoldenService) SetGoldenBuyPrice(price float64) error {
	return gs.repo.SetGoldenBuyPrice(price)
}

func (gs *GoldenService) SetGoldenSellPrice(price float64) error {
	return gs.repo.SetGoldenSellPrice(price)
}
