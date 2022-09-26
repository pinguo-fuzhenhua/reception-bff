package server

import (
	"context"
	"net/http"
	"reflect"

	"time"

	kerr "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pinguo-icc/go-base/v2/iuser"
	kierr "github.com/pinguo-icc/kratos-library/v2/ierr"
	klog "github.com/pinguo-icc/kratos-library/v2/log"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/conf"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/render"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Handler struct {
	// 注入其他微服务客户端
}

func grpcConnection(grpcServer string, logger log.Logger, provider trace.TracerProvider) (*grpc.ClientConn, error) {
	clientOpts := []kgrpc.ClientOption{
		kgrpc.WithEndpoint(grpcServer),
		kgrpc.WithTimeout(time.Second * 30),
		kgrpc.WithBalancerName(roundrobin.Name),
		kgrpc.WithMiddleware(
			tracing.Client(tracing.WithTracerProvider(provider)),
			klog.ClinetMiddleware(logger),
			kierr.GRPCClientMiddleware(), //顺序要求，需要先将grpc错误转换为自定义错误 在记录日志
		),
	}
	conn, err := kgrpc.DialInsecure(
		context.TODO(),
		clientOpts...,
	)
	return conn, err
}

func NewHandler(c *conf.Params, logger log.Logger, traceProvider trace.TracerProvider) (h *Handler, cancel func(), err error) {
	conn, err := grpcConnection("", logger, traceProvider)
	if err != nil {
		return
	}
	_ = conn
	h = &Handler{
		//Client:**api.New****Client(conn),
	}

	cancel = func() {
		//conn.Close()
	}

	return
}

func HandlerFunc(handler interface{}) khttp.HandlerFunc {
	setSpanToRsp := func(ctx khttp.Context) {
		span := trace.SpanFromContext(ctx)
		if span != nil {
			ctx.Response().Header().Set("Trace-Id", span.SpanContext().TraceID().String())
		}
	}
	// 使日志中的 operation 字段为请求 path
	trySetOperation := func(ctx context.Context) {
		if tr, ok := transport.FromServerContext(ctx); ok {
			if tr, ok := tr.(khttp.Transporter); ok {
				khttp.SetOperation(ctx, tr.PathTemplate())
			}
		}
	}

	if fn, ok := handler.(func(*RequestContext) (any, error)); ok {
		return func(ctx khttp.Context) error {
			trySetOperation(ctx)

			next := ctx.Middleware(func(ctx2 context.Context, _ any) (any, error) {
				sCtx := khttpWithContext(ctx, ctx2)
				setSpanToRsp(sCtx)

				return fn(sCtx)
			})
			res, err := next(ctx, nil)
			if err != nil {
				return err
			}
			return render.RenderJSON(ctx, res)
		}
	}
	if fn, ok := handler.(func(*RequestContext) error); ok {
		return func(ctx khttp.Context) error {
			trySetOperation(ctx)

			next := ctx.Middleware(func(ctx2 context.Context, _ any) (any, error) {
				sCtx := khttpWithContext(ctx, ctx2)
				setSpanToRsp(sCtx)
				err := fn(sCtx)
				return nil, err
			})

			_, err := next(ctx, nil)
			return err
		}
	}

	panic("invalid handler")
}

// LoginRequired http filter for backend system
func LoginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := iuser.AcquireUserFromCookie(r)
		if err != nil {
			render.LoginRequired(w)
			return
		}

		ctx := iuser.WithUser(r.Context(), user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func HandlerWithInput[T any](handler func(*RequestContext, *T) (any, error)) khttp.HandlerFunc {
	fn := func(ctx *RequestContext) (any, error) {
		var in T
		if err := injectRequestParams(ctx, &in); err != nil {
			return nil, err
		}
		return handler(ctx, &in)
	}
	return HandlerFunc(fn)
}

func HandlerWithCall[IN, OUT any](rpcCall func(context.Context, *IN, ...grpc.CallOption) (*OUT, error), fn ...func(*RequestContext, *IN)) khttp.HandlerFunc {
	handler := func(ctx *RequestContext) (any, error) {
		var in IN
		if err := injectRequestParams(ctx, &in); err != nil {
			return nil, err
		}
		for _, v := range fn {
			v(ctx, &in)
		}
		return rpcCall(ctx, &in)
	}
	return HandlerFunc(handler)
}

func injectRequestParams(ctx *RequestContext, in any) error {
	if ctx.Request().ContentLength > 0 {
		if err := ctx.Bind(&in); err != nil {
			return kerr.BadRequest(err.Error(), "请求参数解析失败")
		}
	}
	if err := ctx.BindVars(&in); err != nil {
		return kerr.BadRequest(err.Error(), "请求参数解析失败")
	}

	rv := reflect.ValueOf(in).Elem()

	// scope
	if scope := rv.FieldByName("Scope"); scope.IsValid() {
		scope.SetString(ctx.Request().Header.Get("X-Pg-Scope"))
	}

	// editor
	if user, ok := iuser.FromContext(ctx); ok {
		if creator := rv.FieldByName("Creator"); creator.IsValid() && creator.Kind() == reflect.String {
			creator.SetString(user.Name)
		}
		if editor := rv.FieldByName("Editor"); editor.IsValid() && editor.Kind() == reflect.String {
			editor.SetString(user.Name)
		}
	}
	return nil
}
