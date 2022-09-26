package v1

import (
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/server"
)

type Context = *server.RequestContext

// RequestScope acquire request scope
func requestScope(ctx khttp.Context) string {
	const key = "X-Pg-Scope"
	return ctx.Request().Header.Get(key)
}
