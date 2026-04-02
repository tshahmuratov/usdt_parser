package grpcserver

import "go.uber.org/fx"

var Module = fx.Module("grpcserver",
	fx.Provide(NewServer),
	fx.Invoke(RegisterLifecycle),
)
