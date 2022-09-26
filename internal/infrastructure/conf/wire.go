package conf

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "App", "Http", "Params", "Trace"),
)
