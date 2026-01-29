package service

import "github.com/ingoxx/stock-backend/internal/domain"

type StockService struct {
	repo domain.StockInfoRepository
}

func NewStockService(repo domain.StockInfoRepository) *StockService {
	return &StockService{repo: repo}
}

func (ss *StockService) GetStockList() ([]*domain.StockInfo, error) {
	return nil, nil
}
