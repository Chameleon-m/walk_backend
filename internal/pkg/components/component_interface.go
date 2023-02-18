package components

import "context"

type ComponentStartInterface interface {
	Start(ctx context.Context) error
}

type ComponentStopInterface interface {
	Stop(ctx context.Context) error
}

type ComponentReadyInterface interface {
	Ready() <-chan struct{}
}

type ComponentInterface interface {
	ComponentStartInterface
	ComponentStopInterface
	ComponentReadyInterface
}
