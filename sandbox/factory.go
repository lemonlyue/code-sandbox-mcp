package sandbox

import (
	"context"
	"fmt"
)

type FactoryOption func(*Factory)

type Factory struct {
	createFunc func(ctx context.Context, config *Config) (Sandbox, error)
}

func NewFactory(opts ...FactoryOption) *Factory {
	f := &Factory{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func WithDockerCreator(creatorFunc func(ctx context.Context, config *Config) (Sandbox, error)) FactoryOption {
	return func(factory *Factory) {
		factory.createFunc = creatorFunc
	}
}

func (f *Factory) Create(ctx context.Context, config *Config) (Sandbox, error) {
	if f.createFunc == nil {
		return nil, fmt.Errorf("no sandbox creator function provided")
	}
	return f.createFunc(ctx, config)
}
