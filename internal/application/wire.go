package application

import (
	"github.com/google/wire"
	v1 "github.com/pinguo-icc/kratos-template-bff/internal/application/v1"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/server"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(v1.Example), "*"),
	wire.Struct(new(RouterDefines), "*"),
	// NewApp,
	wire.Bind(new(server.Register), new(*RouterDefines)),
)
