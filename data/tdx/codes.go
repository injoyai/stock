package tdx

import (
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx/model"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"github.com/robfig/cron/v3"
	"time"
	"xorm.io/xorm"
)

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

func NewCode(hosts []string, filename string, op ...client.Option) (*Code, error) {

	c, err := tdx.DialWith(tdx.NewHostDial(hosts), func(c *client.Client) {
		c.SetRedial()
		c.SetOption(op...)
	})
	if err != nil {
		return nil, err
	}
	c.Wait.SetTimeout(time.Second * 5)

	db, err := sqlite.NewXorm(filename)
	if err != nil {
		return nil, err
	}

	err = db.Sync2(new(model.Code))
	if err != nil {
		return nil, err
	}

	cc := &Code{
		Client: c,
		db:     db,
		Codes:  nil,
	}

	// 每天早上8点更新数据
	task := cron.New(cron.WithSeconds())
	task.AddFunc("0 0 9 * * *", func() {
		err := g.Retry(cc.Update, 3, func(duration time.Duration) time.Duration {
			return time.Minute * 5
		})
		logs.PrintErr(err)
	})
	task.Start()

	return cc, cc.Update()
}

type Code struct {
	*tdx.Client                        //客户端
	db          *xorms.Engine          //数据库实例
	Codes       map[string]*model.Code //股票缓存
}

// GetName 获取股票名称
func (this *Code) GetName(code string) string {
	if v, ok := this.Codes[code]; ok {
		return v.Name
	}
	return "未知"
}

// GetStocks 获取股票代码,不一定全
func (this *Code) GetStocks() []string {
	ls := []string(nil)
	for k, _ := range this.Codes {
		if len(k) == 8 {
			switch k[:2] {
			case "sz":
				if this.IsStock(protocol.ExchangeSZ, k[2:]) {
					ls = append(ls, k)
				}
			case "sh":
				if this.IsStock(protocol.ExchangeSH, k[2:]) {
					ls = append(ls, k)
				}
			}
		}
	}
	return ls
}

// IsStock 是否是股票,不一定完全,通过百度查询
func (this *Code) IsStock(exchange protocol.Exchange, code string) bool {
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

func (this *Code) Update() error {
	codes, err := this.Code(false)
	if err != nil {
		return err
	}
	codeMap := make(map[string]*model.Code)
	for _, code := range codes {
		codeMap[code.Exchange+code.Code] = code
	}
	this.Codes = codeMap
	return nil
}

// Code 更新股票并返回结果
func (this *Code) Code(byDatabase bool) ([]*model.Code, error) {

	//2. 查询数据库所有股票
	list := []*model.Code(nil)
	if err := this.db.Find(&list); err != nil {
		return nil, err
	}

	//如果是从缓存读取,则返回结果
	if byDatabase {
		return list, nil
	}

	mCode := make(map[string]*model.Code, len(list))
	for _, v := range list {
		mCode[v.Code] = v
	}

	//3. 从服务器获取所有股票代码
	insert := []*model.Code(nil)
	update := []*model.Code(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := this.Client.GetCodeAll(exchange)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.List {
			if _, ok := mCode[v.Code]; ok {
				if mCode[v.Code].Name != v.Name {
					mCode[v.Code].Name = v.Name
					update = append(update, model.NewCode(exchange, v))
				}
			} else {
				code := model.NewCode(exchange, v)
				insert = append(insert, code)
				list = append(list, code)
			}
		}
	}

	//4. 插入或者更新数据库
	err := this.db.SessionFunc(func(session *xorm.Session) error {
		for _, v := range insert {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		for _, v := range update {
			if _, err := session.Where("Code=?", v.Code).Cols("Name").Update(v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return list, nil

}
