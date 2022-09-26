package conf

import (
	"time"

	lconfig "github.com/pinguo-icc/kratos-library/v2/conf"
	"github.com/pinguo-icc/kratos-library/v2/trace"
)

type Bootstrap struct {
	App    *App
	Http   *HTTP
	Params *Params
	Trace  *trace.Config
}

type Params struct {
	TemplateSvcAddr string
}

func Load(env string) (*Bootstrap, error) {
	out := new(Bootstrap)
	err := lconfig.Load(env, out, nil)
	return out, err
}

type (
	HTTP struct {
		Address string
		Timeout time.Duration
	}
	App struct {
		Name   string
		Env    string
		Region string // 部署位置： hz(杭州)、sg(新加坡)
	}
)
