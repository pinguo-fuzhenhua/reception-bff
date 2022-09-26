package infrastructure

import (
	"github.com/google/wire"
	"github.com/pinguo-icc/kratos-library/v2/trace"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/conf"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/server"
)

var ProviderSet = wire.NewSet(
	conf.ProviderSet,
	server.NewHttpServer,
	server.NewHandler,
	trace.NewTracerProvider,
)
