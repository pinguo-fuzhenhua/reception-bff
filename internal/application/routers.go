package application

import (
	"github.com/gorilla/mux"
	v1app "github.com/pinguo-icc/kratos-template-bff/internal/application/v1"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type Context = khttp.Context

func PathParam(ctx Context, name string) (val string, ok bool) {
	raws := mux.Vars(ctx.Request())
	val, ok = raws[name]
	return
}

type RouterDefines struct {
	E *v1app.Example
}

func (rd *RouterDefines) RouteRegister(r *khttp.Router) {
	v1 := r.Group("/v1")
	{
		{
			rd.E.Routes(v1)
		}
	}
}
