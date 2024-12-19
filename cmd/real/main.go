package main

import (
	"log"
	"time"

	_ "embed"

	"github.com/injoyai/lorca"
	"github.com/injoyai/stock/data/tdx"
	tdx2 "github.com/injoyai/tdx"
)

//go:embed index.html
var index string

type StockData struct {
	Price     float64 `json:"price"`
	Time      string  `json:"time"`
	Volume    float64 `json:"volume"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

func main() {
	// 初始化通达信客户端
	client, err := tdx.NewReal(tdx2.Hosts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	lorca.Run(&lorca.Config{
		Width:  1024,
		Height: 768,
		Html:   index,
	}, func(app lorca.APP) error {

		// 绑定获取实时股票数据的函数
		app.Bind("getStockData", func() StockData {
			quotes, err := client.GetQuote("sh688688") // 这里需要替换为实际的股票代码
			if err != nil {
				log.Printf("获取股票数据失败: %v", err)
				return StockData{}
			}
			if len(quotes) == 0 {
				return StockData{}
			}
			quote := quotes[0]

			currentPrice := quote.K.Close.Float64()
			yesterdayClose := quote.K.Last.Float64()
			change := currentPrice - yesterdayClose
			changePct := 0.0
			if yesterdayClose > 0 {
				changePct = change / yesterdayClose * 100
			}

			return StockData{
				Price:     currentPrice,
				Time:      time.Now().Format("15:04:05"),
				Volume:    float64(quote.TotalHand),
				Change:    change,
				ChangePct: changePct,
			}
		})

		// 绑定初始化数据的函数
		app.Bind("initStockData", func() map[string]interface{} {
			quotes, err := client.GetQuote("sh688688") // 同样需要替换为实际的股票代码
			if err != nil {
				log.Printf("获取股票数据失败: %v", err)
				return map[string]interface{}{}
			}
			if len(quotes) == 0 {
				return map[string]interface{}{}
			}
			quote := quotes[0]

			return map[string]interface{}{
				"basePrice":      quote.K.Last.Float64(),
				"name":           "阿里巴巴",
				"code":           "BABA",
				"high":           quote.K.High.Float64(),
				"low":            quote.K.Low.Float64(),
				"open":           quote.K.Open.Float64(),
				"yesterdayClose": quote.K.Last.Float64(),
				"volume":         quote.TotalHand,
				"market":         quote.Amount,
			}
		})

		return nil
	})
}
