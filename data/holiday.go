package data

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/str"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/logs"
	"strings"
	"time"
)

var Holiday = &holiday{m: maps.NewSafe()}

func init() {
	err := Holiday.init()
	logs.PrintErr(err)
}

type holiday struct {
	year int
	m    *maps.Safe
}

func (this *holiday) Is(date string, countries ...string) (bool, error) {
	t, err := time.Parse("20060102", date)
	if err != nil {
		return false, err
	}

	//周末
	if t.Weekday() == 0 || t.Weekday() == 6 {
		return true, nil
	}

	country := "中国"
	if len(countries) > 0 {
		country = countries[0]
	}

	_, month, day := t.Date()
	switch {
	case (country == "中国" && month == 1 && day == 1) ||
		(country == "中国" && month == 10 && day >= 1 && day <= 7) ||
		(country == "中国" && month == 5 && day >= 1 && day <= 3): //五一调休不一定从1-5号
	}

	now := time.Now()
	if t.Year() != now.Year() {
		return false, errors.New("只能查询今年的节假日")
	}

	if this.year != now.Year() {
		if err := this.init(); err != nil {
			return false, err
		}
	}

	m, ok := this.m.Get(country)
	if ok {
		_, ok = m.(*maps.Safe).Get(date)
	}
	return ok, nil
}

func (this *holiday) init() error {

	resp := http.Url("https://www.tdx.com.cn/url/holiday/").Get()
	if resp.Err() != nil {
		return resp.Err()
	}

	bs, err := str.GbkToUtf8(resp.GetBodyBytes())
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	doc.Find("#data").Each(func(i int, selection *goquery.Selection) {
		for _, v := range strings.Split(selection.Text(), "\n") {
			list := strings.Split(v, "|")
			if len(list) == 6 {
				if this.year == 0 && len(list[0]) == 8 {
					this.year = conv.Int(list[0][:4])
				}
				m, _ := this.m.GetOrSetByHandler(list[2], func() (interface{}, error) {
					return maps.NewSafe(), nil
				})
				m.(*maps.Safe).Set(list[0], true)
			}
		}
	})

	return nil
}
