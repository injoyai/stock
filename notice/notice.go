package notice

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/conv/codec"
	"github.com/injoyai/logs"
	"github.com/injoyai/notice/input/forbidden"
	"github.com/injoyai/notice/input/http"
	in_tcp "github.com/injoyai/notice/input/tcp"
	"github.com/injoyai/notice/output/desktop"
	"github.com/injoyai/notice/output/sms"
	"github.com/injoyai/notice/output/tcp"
	"github.com/injoyai/notice/output/wechat"
	"github.com/injoyai/notice/user"
)

var DataDir = "./"

func init() {
	cfg.Init(cfg.WithFile("./config/config.yaml", codec.Yaml))
}

func Init(tcpPort, httpPort int) {

	//加载违禁词规则
	forbidden.Init()

	//加载短信
	sms.Init()

	//加载用户
	logs.PanicErr(user.Init(DataDir))

	//加载微信通知
	logs.PanicErr(wechat.Init(DataDir))

	//加载桌面端通知
	desktop.Init()

	//加载tcp服务
	go tcp.Init(tcpPort, in_tcp.DealMessage)

	//加载http服务
	http.Init(httpPort)
}
