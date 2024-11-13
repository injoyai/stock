package tdx

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/tdx/protocol"
)

type Code struct {
	ID       int64  `json:"id"`                      //主键
	Name     string `json:"name"`                    //名称
	Code     string `json:"code" xorm:"index"`       //代码
	Exchange string `json:"exchange" xorm:"index"`   //交易所
	EditDate int64  `json:"editDate" xorm:"updated"` //修改时间
	InDate   int64  `json:"inDate" xorm:"created"`   //创建时间
}

func NewKline(code string, kline *protocol.Kline) *Kline {
	return &Kline{
		Exchange: code[:2],
		Code:     code[2:],
		Unix:     kline.Time.Unix(),
		Year:     kline.Time.Year(),
		Month:    int(kline.Time.Month()),
		Day:      kline.Time.Day(),
		Hour:     kline.Time.Hour(),
		Minute:   kline.Time.Minute(),
		Open:     kline.Open.Float64(),
		High:     kline.High.Float64(),
		Low:      kline.Low.Float64(),
		Close:    kline.Close.Float64(),
		Volume:   int64(kline.Volume),
		Amount:   int64(kline.Amount),
	}
}

type Kline struct {
	ID       int64   `json:"id"`                    //主键
	Exchange string  `json:"exchange" xorm:"index"` //交易所
	Code     string  `json:"code" xorm:"index"`     //代码
	Unix     int64   `json:"unix"`                  //时间戳
	Year     int     `json:"year"`                  //年
	Month    int     `json:"month"`                 //月
	Day      int     `json:"day"`                   //日
	Hour     int     `json:"hour"`                  //时
	Minute   int     `json:"minute"`                //分
	Open     float64 `json:"open"`                  //开盘价
	High     float64 `json:"high"`                  //最高价
	Low      float64 `json:"low"`                   //最低价
	Close    float64 `json:"close"`                 //最新价,对应历史收盘价
	Volume   int64   `json:"volume"`                //成交量
	Amount   int64   `json:"amount"`                //成交额
	InDate   int64   `json:"inDate" xorm:"created"` //创建时间
}

func NewMinuteTrade(code, date string, trace *protocol.HistoryMinuteTrade) *MinuteTrade {
	return &MinuteTrade{
		Exchange: code[:2],
		Code:     code[2:],
		Date:     date,
		Year:     conv.Int(date[:4]),
		Month:    conv.Int(date[4:6]),
		Day:      conv.Int(date[6:8]),
		Hour:     conv.Int(trace.Time[:2]),
		Minute:   conv.Int(trace.Time[3:5]),
		Second:   0,
		Price:    trace.Price.Float64(),
		Volume:   trace.Volume,
		Number:   0,
		Status:   trace.Status,
	}
}

// MinuteTrade 分时成交
type MinuteTrade struct {
	ID       int64   `json:"id"`                    //主键
	Exchange string  `json:"exchange" xorm:"index"` //交易所
	Code     string  `json:"code" xorm:"index"`     //代码
	Date     string  `json:"date" xorm:"index"`     //日期
	Year     int     `json:"year"`                  //年
	Month    int     `json:"month"`                 //月
	Day      int     `json:"day"`                   //日
	Hour     int     `json:"hour"`                  //时
	Minute   int     `json:"minute"`                //分
	Second   int     `json:"second"`                //秒
	Price    float64 `json:"price"`                 //价格
	Volume   int     `json:"volume"`                //成交量
	Number   int     `json:"number"`                //成交笔数
	Status   int     `json:"status"`                //成交状态,0是买，1是卖
}
