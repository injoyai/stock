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
	Code         string                  //股票代码
	TodayKline   []*tdx.StockKline       //今日K线图
	TodayTrace   []*tdx.StockMinuteTrade //今天分时成交
	HistoryKline []any                   //历史k线图
	HistoryTrace []any                   //历史分时成交
}
