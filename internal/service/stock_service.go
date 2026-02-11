package service

import "github.com/ingoxx/stock-backend/internal/domain"

type StockService struct {
	repo domain.StockInfoRepository
}

func NewStockService(repo domain.StockInfoRepository) *StockService {
	return &StockService{repo: repo}
}

func (ss *StockService) GetStockList() ([]*domain.StockInfo, error) {
	return ss.repo.GetStockList()
}

func (ss *StockService) GetStockInfoForDataList(code string) ([]*domain.StockInfoForDate, error) {
	return ss.repo.GetStockInfoForDataList(code)
}
