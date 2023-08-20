package espresso

import (
	"context"
	internal "espresso/internal"
	_ "espresso/modules" // 注册全部模块到mvc
	"github.com/dan-and-dna/minilog"
	"github.com/prometheus/client_golang/prometheus"
)

type App = internal.App

type Option = internal.Option
type Plugin = internal.Plugin
type Handler = internal.Handler

func New(applyOptions ...Option) *App {
	return internal.New(applyOptions...)
}

func Logger(logger *minilog.MiniLog) Option {
	return internal.Logger(logger)
}

func Metric(registry *prometheus.Registry) Option {
	return internal.Metric(registry)
}

func Http(address string) Option {
	return internal.Http(address)
}

func Grpc(address string) Option {
	return internal.Grpc(address)
}

func RuntimeStat(address string) Option {
	return internal.RuntimeStat(address)
}

func PProf(enable bool) Option {
	return internal.PProf(enable)
}

func ModuleContext(f func(ctx context.Context) context.Context) Option {
	return internal.ModuleContext(f)
}

func ModulePlugin(plugin Plugin) Option {
	return internal.ModulePlugin(plugin)
}
