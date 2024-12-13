package main

import (
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/injoyai/stock/data/tdx/model"
	"strings"
)

func main() {

	dir := "./database"

	oss.RangeFileInfo(dir, func(info *oss.FileInfo) (bool, error) {

		name := info.Name()
		if strings.HasPrefix(name, "sz") || strings.HasPrefix(name, "sh") {

			err := func() error {
				db, err := sqlite.NewXorm(info.Filename())
				if err != nil {
					return err
				}
				db.Close()

				//读取近10天的数据
				ls := model.Klines(nil)
				if err = db.Desc("ID").Limit(10).Find(&ls); err != nil {
					return err
				}

				//连续3天涨停,然后今天大阴线 ls.Get(-4).LimitUp() &&
				if ls.Get(-3).LimitUp() &&
					ls.Get(-2).LimitUp() &&
					ls.Get(-1).RiseRate < -9.9 {

					logs.Debug(ls.Get(-1).Code, "符合策略")

				}

				return nil
			}()

			return true, err

		}

		return true, nil
	})

}
