package tdx

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/tdx"
	"time"
)

func newWorkday(c *tdx.Client, db *xorms.Engine) (*workday, error) {
	w := &workday{
		Client: c,
		db:     db,
		cache:  maps.NewBit(),
	}
	return w, w.db.Sync2(new(Workday))
}

type workday struct {
	*tdx.Client
	db    *xorms.Engine
	cache maps.Bit
}

// Update 更新
func (this *workday) Update() error {
	//获取平安银行的日K线,用作历史是否节假日的判断依据
	//判断日K线是否拉取过
	lastWorkday := new(Workday)
	has, err := this.db.Desc("ID").Get(lastWorkday)
	if err != nil {
		return err
	}
	now := time.Now()
	if !has || lastWorkday.Unix < times.IntegerDay(now).Unix() {
		resp, err := this.Client.GetKlineDayAll("sz000001")
		if err != nil {
			return err
		}
		for _, v := range resp.List {
			if unix := v.Time.Unix(); unix > lastWorkday.Unix {
				_, err = this.db.Insert(&Workday{Unix: unix, Date: v.Time.Format("20060102"), Is: true})
				if err != nil {
					return err
				}
				this.cache.Set(uint64(unix), true)
			}
		}
	}
	return nil
}

// Is 是否是工作日
func (this *workday) Is(t time.Time) bool {
	return this.cache.Get(uint64(t.Unix()))
}

func (this *workday) TodayIs() bool {
	return this.Is(time.Now())
}
