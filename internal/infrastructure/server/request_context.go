package server

import (
	"context"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func khttpWithContext(kctx khttp.Context, ctx context.Context) *RequestContext {
	v := &RequestContext{
		kctx,
		make([]context.Context, 0),
	}
	v.AddSubContext(ctx)
	return v
}

type RequestContext struct {
	khttp.Context
	subContexts []context.Context
}

func (s *RequestContext) AddSubContext(ctx context.Context) {
	s.subContexts = append(s.subContexts, ctx)
}

func (s *RequestContext) Value(key interface{}) interface{} {
	v := s.Context.Value(key)
	if v != nil {
		return v
	}
	for _, ctx := range s.subContexts {
		v := ctx.Value(key)
		if v != nil {
			return v
		}
	}
	return nil
}
