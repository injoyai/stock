package strategy

import "github.com/injoyai/stock/data/tdx"

var (
	All = &Manage{}
)

type Manage struct {
	Handler []func(data *Data)
}

func (this *Manage) Register(f ...func(data *Data)) {
	this.Handler = append(this.Handler, f...)
}

func (this *Manage) Do(data *Data) {
	for _, v := range this.Handler {
		v(data)
	}
}

type Data struct {
	Code                 string             //股票代码
	TodayKline           tdx.Klines         //今日K线图
	TodayTrace           []*tdx.MinuteTrade //今天分时成交
	HistoryKlineMinute   tdx.Klines         //历史k线图
	HistoryKline5Minute  tdx.Klines         //历史k线图
	HistoryKline15Minute tdx.Klines         //历史k线图
	HistoryKline30Minute tdx.Klines         //历史k线图
	HistoryKlineHour     tdx.Klines         //历史k线图
	HistoryKlineDay      tdx.Klines         //历史k线图
	HistoryKlineMonth    tdx.Klines         //历史k线图
	HistoryKlineQuarter  tdx.Klines         //历史k线图
	HistoryKlineYear     tdx.Klines         //历史k线图
	HistoryTrace         []any              //历史分时成交
}
