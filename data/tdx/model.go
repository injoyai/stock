package tdx

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/stock/gui"
	"github.com/injoyai/tdx/protocol"
	"math"
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

//// Update 记录更新时间,避免重复更新
//type Update struct {
//	ID            int64 `json:"id"`                    //主键
//	Code          int64 `json:"code"`                  //代码更新时间
//	KlineMinute   int64 `json:"klineMinute"`           //1分钟K线
//	Kline5Minute  int64 `json:"kline5Minute"`          //5分钟K线
//	Kline15Minute int64 `json:"kline15Minute"`         //15分钟K线
//	Kline30Minute int64 `json:"kline30Minute"`         //30分钟K线
//	KlineHour     int64 `json:"klineHour"`             //小时K线
//	KlineDay      int64 `json:"klineDay"`              //日K线
//	KlineWeek     int64 `json:"klineWeek"`             //周K线
//	KlineMonth    int64 `json:"klineMonth"`            //月K线
//	KlineQuarter  int64 `json:"klineQuarter"`          //季K线
//	KlineYear     int64 `json:"klineYear"`             //年K线
//	InDate        int64 `json:"inDate" xorm:"created"` //创建时间
//}
//
//func (this *Update) GetVar(key string) *conv.Var {
//	switch strings.ToLower(key) {
//	case "code":
//		return conv.New(this.Code)
//	case "klineminute":
//		return conv.New(this.KlineMinute)
//	case "kline5minute":
//		return conv.New(this.Kline5Minute)
//	case "kline15minute":
//		return conv.New(this.Kline15Minute)
//	case "kline30minute":
//		return conv.New(this.Kline30Minute)
//	case "klinehour":
//		return conv.New(this.KlineHour)
//	case "klineday":
//		return conv.New(this.KlineDay)
//	case "klineweek":
//		return conv.New(this.KlineWeek)
//	case "klinemonth":
//		return conv.New(this.KlineMonth)
//	case "klinequarter":
//		return conv.New(this.KlineQuarter)
//	case "klineyear":
//		return conv.New(this.KlineYear)
//	default:
//		return conv.Nil()
//	}
//}
//
//func (this *Update) Update(key string) *Update {
//	switch strings.ToLower(key) {
//	case "code":
//		this.Code = time.Now().Unix()
//	case "klineminute":
//		this.KlineMinute = time.Now().Unix()
//	case "kline5minute":
//		this.Kline5Minute = time.Now().Unix()
//	case "kline15minute":
//		this.Kline15Minute = time.Now().Unix()
//	case "kline30minute":
//		this.Kline30Minute = time.Now().Unix()
//	case "klinehour":
//		this.KlineHour = time.Now().Unix()
//	case "klineday":
//		this.KlineDay = time.Now().Unix()
//	case "klineweek":
//		this.KlineWeek = time.Now().Unix()
//	case "klinemonth":
//		this.KlineMonth = time.Now().Unix()
//	case "klinequarter":
//		this.KlineQuarter = time.Now().Unix()
//	case "klineyear":
//		this.KlineYear = time.Now().Unix()
//	}
//	return this
//}

/**/

/**/

func NewKline(code string, kline *protocol.Kline) *Kline {
	return &Kline{
		Exchange:  code[:2],
		Code:      code[2:],
		Unix:      kline.Time.Unix(),
		Year:      kline.Time.Year(),
		Month:     int(kline.Time.Month()),
		Day:       kline.Time.Day(),
		Hour:      kline.Time.Hour(),
		Minute:    kline.Time.Minute(),
		Open:      kline.Open.Float64(),
		High:      kline.High.Float64(),
		Low:       kline.Low.Float64(),
		Close:     kline.Close.Float64(),
		Volume:    kline.Volume,
		RisePrice: kline.RisePrice().Float64(),
		RiseRate:  kline.RiseRate(),
		Amount:    kline.Amount.Int64(),
	}
}

type Kline struct {
	ID        int64   `json:"id"`                    //主键
	Exchange  string  `json:"exchange" xorm:"index"` //交易所
	Code      string  `json:"code" xorm:"index"`     //代码
	Unix      int64   `json:"unix"`                  //时间戳
	Year      int     `json:"year"`                  //年
	Month     int     `json:"month"`                 //月
	Day       int     `json:"day"`                   //日
	Hour      int     `json:"hour"`                  //时
	Minute    int     `json:"minute"`                //分
	Open      float64 `json:"open"`                  //开盘价
	High      float64 `json:"high"`                  //最高价
	Low       float64 `json:"low"`                   //最低价
	Close     float64 `json:"close"`                 //最新价,对应历史收盘价
	Volume    int64   `json:"volume"`                //成交量
	Amount    int64   `json:"amount"`                //成交额
	RisePrice float64 `json:"risePrice"`             //涨跌幅
	RiseRate  float64 `json:"riseRate"`              //涨跌幅度
	InDate    int64   `json:"inDate" xorm:"created"` //创建时间
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

// Chart k线图 实时价格
func (this Klines) Chart(name string) *gui.Chart {
	c := &gui.Chart{
		Labels: make([]string, len(this)),
		Datasets: []*gui.ChartItem{{
			Label: name,
			Data:  make([]float64, len(this)),
		}},
	}
	for i, v := range this {
		c.Labels[i] = time.Unix(v.Unix, 0).Format("15:04")
		c.Datasets[0].Data[i] = v.Close
		if v.Close > c.Max {
			c.Max = v.Close
		}
		if v.Close < c.Min || c.Min == 0 {
			c.Min = v.Close
		}
	}
	c.Max *= 1.02
	c.Min *= 0.98
	return c
}

func (this Klines) ChartDay(last float64, name string) *gui.Chart {
	dayMinute := 60 * 4
	c := &gui.Chart{
		Labels: make([]string, dayMinute),
		Datasets: []*gui.ChartItem{{
			Label: name,
			Data:  make([]float64, len(this)),
		}},
	}

	now := time.Date(2024, 1, 1, 9, 31, 0, 0, time.Local)
	for i := 0; i < dayMinute/2; i++ {
		c.Labels[i] = now.Add(time.Minute * time.Duration(i)).Format("15:04")
	}

	now = time.Date(2024, 1, 1, 13, 0, 0, 0, time.Local)
	for i := 0; i < dayMinute/2; i++ {
		c.Labels[i+dayMinute/2] = now.Add(time.Minute * time.Duration(i)).Format("15:04")
	}

	var sub float64
	for i, v := range this {
		c.Datasets[0].Data[i] = v.Close
		val := math.Abs(v.Close - last)
		if val > sub {
			sub = val
		}
	}

	c.Max = (last + sub) * 1.02
	c.Min = (last - sub) * 0.98

	return c
}

func (this Klines) Len() int {
	return len(this)
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

// Holiday 节假日
type Holiday struct {
	ID      int64  `json:"id"`                    //主键
	Date    string `json:"date"`                  //日期
	Country string `json:"country"`               //国家
	InDate  int64  `json:"inDate" xorm:"created"` //创建时间
}

// Workday 工作日
type Workday struct {
	ID   int64  `json:"id"`
	Unix int64  `json:"unix"`
	Date string `json:"date"`
	Is   bool   `json:"is"`
}

/**/
