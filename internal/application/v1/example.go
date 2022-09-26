package v1

import (
	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/server"
)

type Example struct {
	*server.Handler
}

// router
func (e *Example) Routes(r *khttp.Router) {
	ep := r.Group("/example")
	ep.GET("/", server.HandlerFunc(e.Get))
	ep.POST("/", server.HandlerWithInput(e.Post))
}

func (e *Example) Get(ctx Context) (any, error) {
	return []string{"example", "get"}, nil
}

type ExamplePostRequest struct {
	Name string `json:"name"`
}

func (e *Example) Post(ctx Context, req *ExamplePostRequest) (any, error) {
	if req.Name == "error" {
		return nil, errors.BadRequest("reason", "response message")
	}
	return req.Name, nil
}
