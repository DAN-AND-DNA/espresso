package events

import (
	"context"
	"espresso/pkg/mvc"
	"time"
)

const (
	ModuleCall      = "event_module_call"
	ModuleCallPanic = "event_module_call_panic"
)

type EventModuleCall struct {
	Module   string
	Function string
	CostTime int64
}

func PublishModuleCall(ctx context.Context, m, f string, startTime int64) {
	mvc.Publish(ModuleCall, ctx, EventModuleCall{Module: m, Function: f, CostTime: time.Now().UnixNano() - startTime})
}

type EventModuleCallPanic struct {
	Module   string
	Function string
	Error    error
}

func PublishModuleCallPanic(ctx context.Context, m, f string, err error) {
	mvc.Publish(ModuleCallPanic, ctx, EventModuleCallPanic{Module: m, Function: f, Error: err})
}
