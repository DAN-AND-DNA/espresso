package network

import (
	"context"
	"espresso/pkg/network/internal"
	"github.com/dan-and-dna/minilog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Network = internal.Network
type Option = internal.Option

func PProf(enable bool) Option {
	return internal.PProf(enable)
}

func Logger(logger *minilog.MiniLog) Option {
	return internal.Logger(logger)
}

func Metric(registry *prometheus.Registry) Option {
	return internal.Metric(registry)
}

func Http(address string, mvc gin.HandlerFunc) Option {
	return internal.Http(address, mvc)
}

func ModuleContext(f ...func(ctx context.Context) context.Context) Option {
	return internal.ModuleContext(f...)
}

func New(options ...Option) *Network {
	return internal.New(options...)
}
