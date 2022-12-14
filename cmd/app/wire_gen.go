// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pinguo-icc/kratos-library/v2/trace"
	"github.com/pinguo-icc/kratos-template-bff/internal/application"
	"github.com/pinguo-icc/kratos-template-bff/internal/application/v1"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/conf"
	"github.com/pinguo-icc/kratos-template-bff/internal/infrastructure/server"
)

// Injectors from wire.go:

func initApp(bootstrap *conf.Bootstrap, logger log.Logger) (*kratos.App, func(), error) {
	http := bootstrap.Http
	config := bootstrap.Trace
	tracerProvider := trace.NewTracerProvider(config)
	params := bootstrap.Params
	handler, cleanup, err := server.NewHandler(params, logger, tracerProvider)
	if err != nil {
		return nil, nil, err
	}
	example := &v1.Example{
		Handler: handler,
	}
	routerDefines := &application.RouterDefines{
		E: example,
	}
	httpServer, cleanup2 := server.NewHttpServer(http, logger, tracerProvider, routerDefines)
	app := newApp(logger, httpServer)
	return app, func() {
		cleanup2()
		cleanup()
	}, nil
}
