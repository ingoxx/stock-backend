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

func (ss *StockService) GetStockHistoryData(code, days string) ([]*domain.StockHistoryDate, error) {
	return ss.repo.GetStockHistoryData(code, days)
}

func (ss *StockService) GetStockRealTimeData(data domain.AverageDownData) ([]*domain.StockInfo, error) {
	return ss.repo.GetStockRealTimeData(data)
}

func (ss *StockService) GetStockRealTimeList() ([]*domain.StockInfo, error) {
	return ss.repo.GetStockRealTimeList()
}

func (ss *StockService) GetStockRealTimeListV2() ([]*domain.StockInfo, error) {
	return ss.repo.GetStockRealTimeListV2()
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

func (ss *StockService) GetGoodStocks(industry, days, lookBackDays, price, trend string) ([]*domain.FilterGoodStock, error) {
	return ss.repo.GetGoodStocks(industry, days, lookBackDays, price, trend)
}

func (ss *StockService) FilterGoodStocksHistory(trend string) ([]string, error) {
	return ss.repo.FilterGoodStocksHistory(trend)
}

func (ss *StockService) StockNoticeFsSet(webHook, word string) error {
	return ss.repo.StockNoticeFsSet(webHook, word)
}

func (ss *StockService) CheckStockNoticeFsSetStatus() bool {
	return ss.repo.CheckStockNoticeFsSetStatus()
}

func (ss *StockService) SendFsInfo() error {
	return ss.repo.SendFsInfo()
}

func (ss *StockService) GetStockHistoryDataDateRange(code, start, end string) ([]*domain.StockHistoryDate, error) {
	return ss.repo.GetStockHistoryDataDateRange(code, start, end)
}

func (ss *StockService) UpdateStockHoldings(code string, price float64, quantity int) ([]*domain.StockInfo, error) {
	return ss.repo.UpdateStockHoldings(code, price, quantity)
}

func (ss *StockService) GetShIndexRealTimeData(isRun bool) (*domain.ShIndex, error) {
	return ss.repo.GetShIndexRealTimeData(isRun)
}

func (ss *StockService) GetCapitalInflowData(isRun bool) ([]*domain.CapitalInflow, error) {
	return ss.repo.GetCapitalInflowData(isRun)
}

func (ss *StockService) AddSelfSelectedStock(code string) ([]*domain.StockInfo, error) {
	return ss.repo.AddSelfSelectedStock(code)
}

func (ss *StockService) GetSelfSelectedStockList() ([]*domain.StockInfo, error) {
	return ss.repo.GetSelfSelectedStockList()
}

func (ss *StockService) UpdateSelfSelectedStock(code string) ([]*domain.StockInfo, error) {
	return ss.repo.UpdateSelfSelectedStock(code)
}

func (ss *StockService) SelfSelectedStockDel(code string) error {
	return ss.repo.SelfSelectedStockDel(code)
}

func (ss *StockService) GetAiApiKey() ([]*domain.AiApiKey, error) {
	return ss.repo.GetAiApiKey()
}

func (ss *StockService) SetAiApiKey(sk domain.AiApiKey) ([]*domain.AiApiKey, error) {
	return ss.repo.SetAiApiKey(sk)
}

func (ss *StockService) GetAverageDownList() ([]*domain.AverageDownData, error) {
	return ss.repo.GetAverageDownList()
}

func (ss *StockService) SetAverageDownInfo(ad domain.AverageDownData) ([]*domain.AverageDownData, error) {
	return ss.repo.SetAverageDownInfo(ad)
}

func (ss *StockService) SetStockTriggeringRulesAlerts(rd domain.TriggeringRules) error {
	return ss.repo.SetStockTriggeringRulesAlerts(rd)
}

func (ss *StockService) GetStockTriggeringRulesAlerts() ([]*domain.TriggeringRules, error) {
	return ss.repo.GetStockTriggeringRulesAlerts()
}

func (ss *StockService) SetStockTagging(data domain.StockTaggingData) (domain.StockTaggingData, error) {
	return ss.repo.SetStockTagging(data)
}

func (ss *StockService) GetStockTagging(code string) (domain.StockTaggingData, error) {
	return ss.repo.GetStockTagging(code)
}
