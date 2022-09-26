GOBIN?=go

GOMODCACHE:=$(shell $(GOBIN) env GOMODCACHE)
moduleName:=$(shell head -n 1 go.mod | awk '{print $$2}')
kratosVersion:=$(shell fgrep go-kratos/kratos/v2 go.mod | awk '{print $$2}')

# protoc
protocImport:=-I$(GOMODCACHE)/github.com/go-kratos/kratos@$(kratosVersion) -I./ -Iapi
protocOut:=--go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. 
f:=api/api.proto

# init 
appName:=$(notdir $(shell pwd))
namespace=$(shell head -n 4 ./deploy/base/namespace.yaml | tail -n 1 | awk '{print $$2}')
podName=$(shell head -n 10 ./deploy/base/deploy.yaml | tail -n 1 | awk '{print $$2}')

.PHONY: run proto build wire help

run:
	bin/app -env local
proto:
	# make protoc f=api/foo.proto
	protoc $(protocImport) $(protocOut) $(f)
build:
	$(GOBIN) build  -o ./bin/ $(moduleName)/cmd/app
wire:
	wire gen $(moduleName)/cmd/app

init:
	LC_ALL=C sed -i '' 's/$(subst /,\/,$(moduleName))/github.com\/pinguo-icc\/$(appName)/g' `grep -rl $(subst /,\/,$(moduleName)) ./`
	LC_ALL=C sed -i '' 's/$(namespace)/$(appName)-namespace/g' `grep -rl $(namespace) ./deploy`
	LC_ALL=C sed -i '' 's/$(podName)/$(appName)/g' `grep -rl $(podName) ./deploy`
	LC_ALL=C sed -i '' 's/$(podName)/$(appName)/g' `grep -rl $(podName) ./configs`