package internal

import (
	"context"
	"espresso/pkg/ctxhelper"
	"go.uber.org/zap"
	"time"
)

type HelloRequest struct {
	Content string `json:"content" form:"content" binding:"required"`
}

type HelloResponse struct {
	Content string `json:"content,omitempty"`
}

func (module *Module) Hello(ctx context.Context, request *HelloRequest, response *HelloResponse) error {
	defer ctxhelper.FetchTrace(ctx).Start().Stop()

	logger := ctxhelper.FetchLogger(ctx)

	logger.Info("received a new message", zap.String("clientIp", ctxhelper.FetchClientIP(ctx)), zap.String("content", request.Content))
	response.Content = request.Content
	return nil
}

type TryPanicRequest struct {
}

type TryPanicResponse struct {
}

func (module *Module) TryPanic(ctx context.Context, request *TryPanicRequest, response *TryPanicResponse) error {
	defer ctxhelper.FetchTrace(ctx).Start().Stop()

	panic("try panic")
	return nil
}

type SlowRequest struct {
}

type SlowResponse struct {
}

func (module *Module) Slow(ctx context.Context, request *SlowRequest, response *SlowResponse) error {
	defer ctxhelper.FetchTrace(ctx).Start().Stop()

	time.Sleep(4 * time.Second)
	return nil
}
