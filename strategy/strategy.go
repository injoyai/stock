package strategy

import "github.com/injoyai/stock/data/tdx"

var (
	All = &Manage{}
)

type Manage struct {
	Handler []func(i Interface)
}

func (this *Manage) Register(f ...func(i Interface)) {
	this.Handler = append(this.Handler, f...)
}

func (this *Manage) Do(i Interface) {
	for _, v := range this.Handler {
		v(i)
	}
}

type Data struct {
	Code                 string       //股票代码
	TodayKline           tdx.Klines   //今日K线图
	TodayTrace           []*tdx.Trade //今天分时成交
	HistoryKlineMinute   tdx.Klines   //历史k线图
	HistoryKline5Minute  tdx.Klines   //历史k线图
	HistoryKline15Minute tdx.Klines   //历史k线图
	HistoryKline30Minute tdx.Klines   //历史k线图
	HistoryKlineHour     tdx.Klines   //历史k线图
	HistoryKlineDay      tdx.Klines   //历史k线图
	HistoryKlineMonth    tdx.Klines   //历史k线图
	HistoryKlineQuarter  tdx.Klines   //历史k线图
	HistoryKlineYear     tdx.Klines   //历史k线图
	HistoryTrace         []any        //历史分时成交
}

type Interface interface {
	GetKlineMinute(code string) ([]*tdx.Kline, error)
	GetKline5Minute(code string) ([]*tdx.Kline, error)
	GetKline15Minute(code string) ([]*tdx.Kline, error)
	GetKline30Minute(code string) ([]*tdx.Kline, error)
	GetKlineHour(code string) ([]*tdx.Kline, error)
	GetKlineDay(code string) ([]*tdx.Kline, error)
	GetKlineWeek(code string) ([]*tdx.Kline, error)
	GetKlineMonth(code string) ([]*tdx.Kline, error)
	GetKlineQuarter(code string) ([]*tdx.Kline, error)
	GetKlineYear(code string) ([]*tdx.Kline, error)
	GetTraceMinute(code string) ([]*tdx.Trade, error)
}
