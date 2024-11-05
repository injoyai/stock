package tdx

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

var (
	Tdx *tdx.Client
)

func Init() error {
	var err error
	Tdx, err = tdx.Dial(cfg.GetString("tdx.address"), tdx.WithDebug(false))
	if err != nil {
		return err
	}

	return common.DB.Sync2(new(StockCode))

}

// UpdateCode 更新数据库的股票代码
func UpdateCode(refresh bool) error {
	exchanges := []protocol.Exchange{
		protocol.ExchangeSH,
		protocol.ExchangeSZ,
		//protocol.ExchangeBJ,
	}

	for _, exchange := range exchanges {

		resp, err := Tdx.GetCodeAll(exchange)
		if err != nil {
			return err
		}
		logs.Debug(resp.Count)

		list := []*StockCode(nil)
		if err := common.DB.Find(&list); err != nil {
			return err
		}
		m := map[string]*StockCode{}
		for _, v := range list {
			m[v.Code] = v
		}

		for _, v := range resp.List {
			if _, ok := m[v.Code]; ok {
				if refresh {
					if _, err := common.DB.Where("Code=?", v.Code).Cols("Name").Update(&StockCode{Name: v.Name}); err != nil {
						return err
					}
				}
				continue
			}
			_, err := common.DB.Insert(&StockCode{
				Name:     v.Name,
				Code:     v.Code,
				Exchange: exchange.String(),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func UpdateHistoryKline() error {

	return nil
}

// GetStockHistoryKline 历史K线
func GetStockHistoryKline(Type TypeKline, code string) error {
	resp, err := Tdx.GetKlineAll(Type.Uint8(), code)
	if err != nil {
		return err
	}

	for _, v := range resp.List {
		year, month, day := v.Time.Date()
		x := &StockKline{
			Exchange: code[:2],
			Code:     code[2:],
			Year:     year,
			Month:    int(month),
			Day:      day,
			Hour:     v.Time.Hour(),
			Minute:   v.Time.Minute(),
			Open:     v.Open.Float64(),
			High:     v.High.Float64(),
			Low:      v.Low.Float64(),
			Latest:   v.Close.Float64(),
			Volume:   v.Volume,
			Amount:   v.Amount,
		}
		if err := common.TSDB.Write(Type.TableName(), x.Tags(), x.GMap()); err != nil {
			return err
		}
	}

	return nil
}
