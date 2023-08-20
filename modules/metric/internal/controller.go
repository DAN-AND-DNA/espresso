package internal

import (
	"context"
	"espresso/pkg/ctxhelper"
	"espresso/pkg/events"
	"espresso/pkg/mvc"
	"go.uber.org/zap"
)

type Controller struct {
	model *Model
}

func (ctr *Controller) Init(model *Model) error {
	ctr.model = model

	// 注册事件处理
	_ = mvc.Subscribe(events.ModuleCall, ctr.OnModuleCall)
	_ = mvc.Subscribe(events.ModuleCallPanic, ctr.OnModuleCallPanic)

	return nil
}

func (ctr *Controller) Clean() {
	// 解绑事件处理
	_ = mvc.UnSubscribe(events.ModuleCall, ctr.OnModuleCall)
	_ = mvc.UnSubscribe(events.ModuleCallPanic, ctr.OnModuleCallPanic)

}

func (ctr *Controller) OnModuleCall(ctx context.Context, event events.EventModuleCall) {
	requestId := ctxhelper.FetchRequestId(ctx)
	logger := ctxhelper.FetchLogger(ctx)

	// 大于3秒，打印堆栈
	if event.CostTime/1000000000 >= 3 {
		t := ctxhelper.FetchTrace(ctx)
		traceInfo := ""
		if t != nil {
			traceInfo = t.String()
		}
		logger.Warn("slow call",
			zap.String("module", event.Module),
			zap.String("function", event.Function),
			zap.String("requestId", requestId),
			zap.String("trace", traceInfo))

		ctr.model.MetricSlowCall(event.Module, event.Function)
	}

	// 调用统计次数
	ctr.model.MetricCall(event.Module, event.Function)
}

func (ctr *Controller) OnModuleCallPanic(ctx context.Context, event events.EventModuleCallPanic) {
	requestId := ctxhelper.FetchRequestId(ctx)
	logger := ctxhelper.FetchLogger(ctx)

	logger.Error("module panic", zap.String("module", event.Module), zap.String("function", event.Function), zap.Error(event.Error), zap.String("requestId", requestId))

	ctr.model.MetricPanicCall(event.Module, event.Function)
}
