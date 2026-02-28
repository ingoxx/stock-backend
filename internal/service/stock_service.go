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

func (ss *StockService) GetStockInfoForDataList(code string) ([]*domain.StockHistoryDate, error) {
	return ss.repo.GetStockInfoForDataList(code)
}

func (ss *StockService) GetStockIndustryList() ([]*domain.StockIndustryMap, error) {
	return ss.repo.GetStockIndustryList()
}

func (ss *StockService) GetIndustryStockUpDown() ([]*domain.StockIndustryUpDown, error) {
	return ss.repo.GetIndustryStockUpDown()
}

func (ss *StockService) GetStockMarketData() (domain.StockMarketData, error) {
	return ss.repo.GetStockMarketData()
}

func (ss *StockService) GetStockDataSwitch() error {
	return ss.repo.GetStockDataSwitch()
}

func (ss *StockService) GetStockDataStatus() error {
	return ss.repo.GetStockDataStatus()
}

func (ss *StockService) GetIndustryData(name string) ([]*domain.StockInfo, error) {
	return ss.repo.GetIndustryData(name)
}

func (ss *StockService) GetStockHistoryData(code string) ([]*domain.StockHistoryDate, error) {
	return ss.repo.GetStockHistoryData(code)
}
