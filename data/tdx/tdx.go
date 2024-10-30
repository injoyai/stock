package tdx

import (
	"errors"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/common"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"strings"
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
		protocol.ExchangeBJ,
	}

	for _, exchange := range exchanges {

		resp, err := Tdx.GetStockAll(exchange)
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
	exchange, code, err := GetExchangeCode(code)
	if err != nil {
		return err
	}
	resp, err := Tdx.GetStockKlineAll(protocol.TypeKline(Type), exchange, code)
	if err != nil {
		return err
	}

	for _, v := range resp.List {
		year, month, day := v.Time.Date()
		x := &StockKline{
			Exchange: exchange.String(),
			Code:     code,
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

func GetExchangeCode(code string) (protocol.Exchange, string, error) {
	if len(code) != 8 {
		return 0, "", errors.New("代码错误,例sh000001")
	}
	code = strings.ToLower(code)
	var exchange protocol.Exchange
	switch code[:2] {
	case "sh":
		exchange = protocol.ExchangeSH
	case "sz":
		exchange = protocol.ExchangeSZ
	case "bj":
		exchange = protocol.ExchangeBJ
	default:
		return 0, "", errors.New("无效交易所编号,例sh000001")
	}
	return exchange, code[2:], nil
}
