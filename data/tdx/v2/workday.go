package tdx

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/times"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/robfig/cron/v3"
	"time"
	"xorm.io/xorm"
)

func NewWorkday(hosts []string, filename string, op ...client.Option) (*workday, error) {

	c, err := tdx.DialWith(tdx.NewHostDial(hosts), op...)
	if err != nil {
		return nil, err
	}

	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}

	if err := db.Sync2(new(Workday)); err != nil {
		return nil, err
	}

	w := &workday{
		Client: c,
		db:     db,
		cache:  maps.NewBit(),
	}

	// 每天早上8点更新数据
	cron.New(cron.WithSeconds()).AddFunc("0 0 8 * * *", func() {
		err := w.Update()
		logs.PrintErr(err)
	})

	return w, w.Update()
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

		this.db.SessionFunc(func(session *xorm.Session) error {
			for _, v := range resp.List {
				if unix := v.Time.Unix(); unix > lastWorkday.Unix {
					_, err = session.Insert(&Workday{Unix: unix, Date: v.Time.Format("20060102"), Is: true})
					if err != nil {
						return err
					}
					this.cache.Set(uint64(unix), true)
				}
			}
			return nil
		})

	}
	return nil
}

// Is 是否是工作日
func (this *workday) Is(t time.Time) bool {
	return this.cache.Get(uint64(t.Unix()))
}

// TodayIs 今天是否是工作日
func (this *workday) TodayIs() bool {
	return this.Is(time.Now())
}
