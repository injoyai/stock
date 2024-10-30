package tdx

type StockCode struct {
	ID       int64  `json:"id"`                      //主键
	Name     string `json:"name"`                    //名称
	Code     string `json:"code" xorm:"index"`       //代码
	Exchange string `json:"exchange" xorm:"index"`   //交易所
	EditDate int64  `json:"editDate" xorm:"updated"` //修改时间
	InDate   int64  `json:"inDate" xorm:"created"`   //创建时间
}

type StockKline struct {
	Exchange string  `json:"exchange" xorm:"index"` //交易所
	Code     string  `json:"code" xorm:"index"`     //代码
	Year     int     `json:"year"`                  //年
	Month    int     `json:"month"`                 //月
	Day      int     `json:"day"`                   //日
	Hour     int     `json:"hour"`                  //时
	Minute   int     `json:"minute"`                //分
	Open     float64 `json:"open"`                  //开盘价
	High     float64 `json:"high"`                  //最高价
	Low      float64 `json:"low"`                   //最低价
	Latest   float64 `json:"close"`                 //最新价,对应历史收盘价
	Volume   float64 `json:"volume"`                //成交量
	Amount   float64 `json:"amount"`                //成交额
}

func (this *StockKline) Tags() map[string]string {
	return map[string]string{
		"exchange": this.Exchange,
	}
}

func (this *StockKline) GMap() map[string]any {
	return map[string]any{
		"exchange": this.Exchange,
		"code":     this.Code,
		"year":     this.Year,
		"month":    this.Month,
		"day":      this.Day,
		"hour":     this.Hour,
		"minute":   this.Minute,
		"open":     this.Open,
		"high":     this.High,
		"low":      this.Low,
		"close":    this.Latest,
		"volume":   this.Volume,
		"amount":   this.Amount,
	}
}

// StockMinuteTrade 分时成交
type StockMinuteTrade struct {
	ID       int64   `json:"id"`
	Exchange string  `json:"exchange" xorm:"index"` //交易所
	Code     string  `json:"code" xorm:"index"`     //代码
	Year     int     `json:"year"`                  //年
	Month    int     `json:"month"`                 //月
	Day      int     `json:"day"`                   //日
	Hour     int     `json:"hour"`                  //时
	Minute   int     `json:"minute"`                //分
	Price    float64 `json:"price"`                 //价格
	Volume   int64   `json:"volume"`                //成交量
	Number   int     `json:"number"`                //成交笔数
	Status   int     `json:"status"`                //成交状态,0是买，1是卖
}
