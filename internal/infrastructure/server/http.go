package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/encoding/json"
	kerr "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/pinguo-icc/go-base/v2/ierr"
	klog "github.com/pinguo-icc/kratos-library/v2/log"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/conf"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/render"
	"go.opentelemetry.io/otel/trace"
)

type Register interface {
	RouteRegister(*khttp.Router)
}

// New new a bm server.
func NewHttpServer(config *conf.HTTP, logger log.Logger, provider trace.TracerProvider, r Register) (*khttp.Server, func()) {
	loggerWithMethod := log.With(
		logger,
		"method",
		log.Valuer(func(ctx context.Context) interface{} {
			if c, ok := ctx.(khttp.Context); ok {
				return c.Request().Method
			}
			return ""
		}),
		"path",
		log.Valuer(func(ctx context.Context) interface{} {
			if c, ok := ctx.(khttp.Context); ok {
				return c.Request().URL.Path
			}
			return ""
		}),
	)

	var opts = []khttp.ServerOption{
		khttp.Logger(logger),
		khttp.Address(config.Address),
		khttp.Timeout(config.Timeout),
		khttp.Middleware(
			recovery.Recovery(recovery.WithLogger(loggerWithMethod)),
			tracing.Server(tracing.WithTracerProvider(provider)), // 顺序有要求, 需要在logger前, 否则logger不能获取span/trace id
			klog.ServerMiddleware(loggerWithMethod),
		),
		khttp.ErrorEncoder(buildErrorEncoder(logger)),
		khttp.Filter(
			cors(),
		),
	}

	svc := khttp.NewServer(opts...)
	route := svc.Route("/")
	registerRouter(route)
	r.RouteRegister(route)

	cancelFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		svc.Shutdown(ctx)
	}

	return svc, cancelFn
}

func cors() khttp.FilterFunc {
	return handlers.CORS(
		handlers.AllowedOriginValidator(func(s string) bool {
			return strings.Contains(s, "camera360.com")
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "HEAD", "PUT", "PATCH", "DELETE"}),
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{
			"DNT", "X-CustomHeader", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Authorization",
		}),
	)
}

var errCodec = encoding.GetCodec(json.Name)

func buildErrorEncoder(logger log.Logger) khttp.EncodeErrorFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var httpCode int
		resp := new(render.ErrorJSON)
		if ie, ok := ierr.FromError(err); ok {
			resp.Code = ie.SubCode
			resp.Message = ie.Message
			httpCode = ie.Code
		} else {
			se := kerr.FromError(err)
			resp.Code = int(se.Code)
			resp.Message = se.Message
			httpCode = int(se.Code)
		}
		body, merr := errCodec.Marshal(resp)
		if merr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithContext(r.Context(), logger).Log(log.LevelError, "b_err", err, "e_err", merr)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if httpCode >= 1000 {
			httpCode = ierr.CustomizeCode
		}
		w.WriteHeader(int(httpCode))
		wlen, werr := w.Write(body)
		log.WithContext(r.Context(), logger).Log(log.LevelWarn, "b_err", err, "w_len", wlen, "w_err", werr)
	}
}
