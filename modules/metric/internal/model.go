package internal

import (
	"context"
	"errors"
	"espresso/pkg/ctxhelper"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	ErrorEmptyPrometheusRegistry = errors.New("empty prometheus registry")
)

type Model struct {
	callMetric      *prometheus.CounterVec
	slowCallMetric  *prometheus.CounterVec
	panicCallMetric *prometheus.CounterVec
}

func (model *Model) Init(ctx context.Context) error {
	logger := ctxhelper.FetchLogger(ctx)
	registry := ctxhelper.FetchRegistry(ctx)
	if registry == nil {
		logger.Error("model init", zap.Error(ErrorEmptyPrometheusRegistry))
		return ErrorEmptyPrometheusRegistry
	} else {
		model.callMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mvc",
			Name:      "module_call",
			Help:      "模块调用数",
		}, []string{"module", "function"})

		model.slowCallMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mvc",
			Name:      "slow_module_call",
			Help:      "模块慢查询数",
		}, []string{"module", "function"})

		model.panicCallMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mvc",
			Name:      "panic_module_panic_call",
			Help:      "模块调用崩溃数",
		}, []string{"module", "function"})

		if err := registry.Register(model.callMetric); err != nil {
			logger.Error("model init", zap.Error(err))
			return err
		}

		if err := registry.Register(model.slowCallMetric); err != nil {
			logger.Error("model init", zap.Error(err))
			return err
		}

		if err := registry.Register(model.panicCallMetric); err != nil {
			logger.Error("model init", zap.Error(err))
			return err
		}
	}
	return nil
}

func (model *Model) MetricSlowCall(module, function string) {
	if model.slowCallMetric == nil {
		return
	}
	model.slowCallMetric.WithLabelValues(module, function).Add(1)
}

func (model *Model) MetricCall(module, function string) {
	if model.callMetric == nil {
		return
	}
	model.callMetric.WithLabelValues(module, function).Add(1)
}

func (model *Model) MetricPanicCall(module, function string) {
	if model.panicCallMetric == nil {
		return
	}
	model.panicCallMetric.WithLabelValues(module, function).Add(1)
}
