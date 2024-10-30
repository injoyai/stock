package tdx

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/tdx"
)

var (
	TDX *tdx.Client
)

func Init() error {
	var err error
	TDX, err = tdx.Dial(cfg.GetString("tdx.address"))
	if err != nil {
		return err
	}

	//common.DB.Sync2()

	return nil
}
