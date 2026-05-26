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

func (ss *StockService) GetStockInfoData(code string) (*domain.StockInfo, error) {
	return ss.repo.GetStockInfoData(code)
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

func (ss *StockService) GetStockRealTimeData(code, price, hold string) ([]*domain.StockInfo, error) {
	return ss.repo.GetStockRealTimeData(code, price, hold)
}

func (ss *StockService) GetStockRealTimeList() ([]*domain.StockInfo, error) {
	return ss.repo.GetStockRealTimeList()
}

func (ss *StockService) DelSelfSelectedStock(code string) error {
	return ss.repo.DelSelfSelectedStock(code)
}

func (ss *StockService) UpdateStockDealStatus(code string, status int) ([]*domain.StockInfo, error) {
	return ss.repo.UpdateStockDealStatus(code, status)
}

func (ss *StockService) GetHistoryTradeDataList() ([]*domain.StockInfo, error) {
	return ss.repo.GetHistoryTradeDataList()
}

func (ss *StockService) AddHistoryTradeData(code string, TradeType int) ([]*domain.StockInfo, error) {
	return ss.repo.AddHistoryTradeData(code, TradeType)
}

func (ss *StockService) StockRealTimeInfoSwitch(status int) (string, error) {
	return ss.repo.StockRealTimeInfoSwitch(status)
}

func (ss *StockService) GetStockRtData(code string) (*domain.StockInfo, error) {
	return ss.repo.GetStockRtData(code)
}

func (ss *StockService) StockNoticeSwitch(status int) (int, error) {
	return ss.repo.StockNoticeSwitch(status)
}

func (ss *StockService) GetGoodStocks(industry, days, lookBackDays string) ([]*domain.FilterGoodStock, error) {
	return ss.repo.GetGoodStocks(industry, days, lookBackDays)
}

func (ss *StockService) FilterGoodStocksHistory() ([]string, error) {
	return ss.repo.FilterGoodStocksHistory()
}
