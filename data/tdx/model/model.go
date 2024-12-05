package model

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/tdx/protocol"
	"time"
)

// IsStock 是否是股票,不一定完全,通过百度查询
func IsStock(exchange protocol.Exchange, code string) bool {
	switch {
	case exchange == protocol.ExchangeSH &&
		(code[0:1] == PrefixSHA ||
			code[0:3] == PrefixSH900):
		return true

	case exchange == protocol.ExchangeSZ &&
		(code[0:1] == PrefixSZChinext ||
			code[0:3] == PrefixSZ000 ||
			code[0:3] == PrefixSZ002 ||
			code[0:3] == PrefixSZ200):
		return true
	}
	return false
}

const (
	PrefixSHA   = "6"   //沪市A股
	PrefixSH900 = "900" //沪市B股
	PrefixSH58  = "58"  //沪市权证

	PrefixSZChinext = "3"   //创业板
	PrefixSZ000     = "000" //深市主板
	PrefixSZ002     = "002" //深市中小板
	PrefixSZWarrant = "03"  //深市权证
	PrefixSZ200     = "200" //深市B股
)

func NewCode(exchange protocol.Exchange, v *protocol.Code) *Code {
	return &Code{
		Name:     v.Name,
		Code:     v.Code,
		Exchange: exchange.String(),
		Stock:    IsStock(exchange, v.Code),
	}
}

type Code struct {
	ID       int64  `json:"id"`                      //主键
	Name     string `json:"name"`                    //名称
	Code     string `json:"code" xorm:"index"`       //代码
	Exchange string `json:"exchange" xorm:"index"`   //交易所
	Stock    bool   `json:"stock" xorm:"index"`      //是否是股票
	EditDate int64  `json:"editDate" xorm:"updated"` //修改时间
	InDate   int64  `json:"inDate" xorm:"created"`   //创建时间
}

/**/

type Update struct {
	ID            int64  `json:"id"`            //
	Name          string `json:"name"`          //股票名称
	KlineMinute   int64  `json:"klineMinute"`   //分时K线更新时间
	Kline5Minute  int64  `json:"kline5Minute"`  //5分钟K线更新时间
	Kline15Minute int64  `json:"kline15Minute"` //15分钟K线更新时间
	Kline30Minute int64  `json:"kline30Minute"` //30分钟K线更新时间
	KlineHour     int64  `json:"klineHour"`     //小时K线更新时间
	KlineDay      int64  `json:"klineDay"`      //日K线更新时间
	KlineWeek     int64  `json:"klineWeek"`     //周K线更新时间
	KlineMonth    int64  `json:"klineMonth"`    //月K线更新时间
	KlineQuarter  int64  `json:"klineQuarter"`  //季K线更新时间
	KlineYear     int64  `json:"klineYear"`     //年K线更新时间
	Trade         int64  `json:"trade"`         //分时成交更新时间
}

/**/

/**/

func NewKline(code string, kline *protocol.Kline, node time.Time) *Kline {
	return &Kline{
		Exchange:  code[:2],
		Code:      code[2:],
		Node:      node.Unix(),
		Unix:      kline.Time.Unix(),
		Year:      kline.Time.Year(),
		Month:     int(kline.Time.Month()),
		Day:       kline.Time.Day(),
		Hour:      kline.Time.Hour(),
		Minute:    kline.Time.Minute(),
		Last:      kline.Last.Float64(),
		Open:      kline.Open.Float64(),
		High:      kline.High.Float64(),
		Low:       kline.Low.Float64(),
		Close:     kline.Close.Float64(),
		Volume:    kline.Volume,
		RisePrice: kline.RisePrice().Float64(),
		RiseRate:  kline.RiseRate(),
		Amount:    kline.Amount.Float64(),
	}
}

type Kline struct {
	ID        int64   `json:"id"`                    //主键
	Exchange  string  `json:"exchange" xorm:"index"` //交易所
	Code      string  `json:"code" xorm:"index"`     //代码
	Node      int64   `json:"node" xorm:"index"`     //时间节点
	Unix      int64   `json:"unix"`                  //时间戳
	Year      int     `json:"year"`                  //年
	Month     int     `json:"month"`                 //月
	Day       int     `json:"day"`                   //日
	Hour      int     `json:"hour"`                  //时
	Minute    int     `json:"minute"`                //分
	Last      float64 `json:"last"`                  //昨收
	Open      float64 `json:"open"`                  //开盘价
	High      float64 `json:"high"`                  //最高价
	Low       float64 `json:"low"`                   //最低价
	Close     float64 `json:"close"`                 //最新价,对应历史收盘价
	Volume    int64   `json:"volume"`                //成交量
	Amount    float64 `json:"amount"`                //成交额
	RisePrice float64 `json:"risePrice"`             //涨跌幅
	RiseRate  float64 `json:"riseRate"`              //涨跌幅度
	InDate    int64   `json:"inDate" xorm:"created"` //创建时间
}

// LimitUp 是否涨停
func (this *Kline) LimitUp() bool {

	return false
}

func NewKlineTable(suffix string) *KlineTable {
	return &KlineTable{
		tableName: "Kline" + suffix,
	}
}

type KlineTable struct {
	Kline     `xorm:"extends"`
	tableName string
}

func (this *KlineTable) TableName() string {
	return this.tableName
}

type KlineChart struct {
	Time  []int64   `json:"time"`
	Price []float64 `json:"price"`
}

type Klines []*Kline

func (this Klines) Len() int {
	return len(this)
}

func (this Klines) Get(index int) *Kline {
	if index < 0 && -index <= len(this) {
		return this[len(this)+index]
	} else if index >= 0 && index < len(this) {
		return this[index]
	}
	return nil
}

// Avg k线的平均值
func (this Klines) Avg(num int) (float64, float64, float64, float64) {
	ls := this
	if len(this) > num {
		ls = this[len(this)-num:]
	}
	var totalOpen, totalHigh, totalLow, totalClose float64
	for _, v := range ls {
		totalOpen += v.Open
		totalHigh += v.High
		totalLow += v.Low
		totalClose += v.Close
	}
	return totalOpen / float64(num), totalHigh / float64(num), totalLow / float64(num), totalClose / float64(num)
}

// AvgClose 收盘平均线
func (this Klines) AvgClose(num int) float64 {
	_, _, _, _close := this.Avg(num)
	return _close
}

// AvgClose5 5条收盘平均线
func (this Klines) AvgClose5() float64 { return this.AvgClose(5) }

// AvgClose10 10条收盘平均线
func (this Klines) AvgClose10() float64 { return this.AvgClose(10) }

// AvgClose30 30条收盘平均线
func (this Klines) AvgClose30() float64 { return this.AvgClose(30) }

// RiseRate 首尾涨幅
func (this Klines) RiseRate() float64 {
	if len(this) < 2 {
		return 0
	}
	return RiseRate(this[0], this[len(this)-1])
}

// RiseRate 涨幅度
func RiseRate(k1, k2 *Kline) float64 {
	return (k2.Close - k1.Open) / k1.Open
}

/**/

func NewTrade(code, date string, trace *protocol.HistoryMinuteTrade) *Trade {
	return &Trade{
		Exchange: code[:2],
		Code:     code[2:],
		Date:     date,
		Hour:     conv.Int8(trace.Time[:2]),
		Minute:   conv.Int8(trace.Time[3:5]),
		Second:   0,
		Price:    trace.Price.Float64(),
		Volume:   trace.Volume,
		Number:   0,
		Status:   int8(trace.Status),
	}
}

// Trade 分时成交
type Trade struct {
	ID       int64   `json:"id"`                    //主键
	Exchange string  `json:"exchange" xorm:"index"` //交易所
	Code     string  `json:"code" xorm:"index"`     //代码
	Date     string  `json:"date" xorm:"index"`     //日期
	Hour     int8    `json:"hour"`                  //时
	Minute   int8    `json:"minute"`                //分
	Second   int8    `json:"second"`                //秒,通达信没有秒
	Price    float64 `json:"price"`                 //价格
	Volume   int     `json:"volume"`                //成交量
	Number   int     `json:"number"`                //成交笔数
	Status   int8    `json:"status"`                //成交状态,0是买，1是卖
}

/**/

// Workday 工作日
type Workday struct {
	ID   int64  `json:"id"`
	Unix int64  `json:"unix"`
	Date string `json:"date"`
	Is   bool   `json:"is"`
}

/**/
