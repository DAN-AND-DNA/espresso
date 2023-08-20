package internal

import "context"

type Module interface {
	ModuleUID() string
	ModuleInit(ctx context.Context) error
	ModuleExit(ctx context.Context) error
	ModuleClean(ctx context.Context) error
}
