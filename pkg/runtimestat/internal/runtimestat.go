package internal

import (
	"context"
	"espresso/pkg/ctxhelper"
	"espresso/pkg/gosafe"
	"github.com/arl/statsviz"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type options struct {
	address string
}

type Option func(opts *options)

type RuntimeStat struct {
	opts *options
	srv  *http.Server
}

func Address(address string) Option {
	return func(opts *options) {
		opts.address = address
	}
}

func New(applyOptions ...Option) *RuntimeStat {
	runtimeStat := &RuntimeStat{
		opts: &options{},
	}

	for _, applyOption := range applyOptions {
		applyOption(runtimeStat.opts)
	}

	runtimeStat.srv = &http.Server{
		Addr: runtimeStat.opts.address,
	}

	return runtimeStat
}

func (rs *RuntimeStat) Run(ctx context.Context) error {
	logger := ctxhelper.FetchLogger(ctx)

	// 注册路由
	if err := statsviz.RegisterDefault(); err != nil {
		logger.Error("service run", zap.Error(err), zap.String("service", "runtimeStat"), zap.Bool("result", false))
		return err
	}

	gosafe.GoSafe(ctx,
		func(ctx context.Context) {
			logger.Info("service run", zap.String("listenAddress", rs.opts.address), zap.Bool("result", true), zap.String("service", "runtimeStat"))
			if err := rs.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("service run", zap.Error(err), zap.String("service", "runtimeStat"), zap.Bool("result", false))
				return
			}
		},
		func(ctx context.Context, err error) {
			if err != nil {
				logger.Error("service exit", zap.Error(err), zap.String("service", "runtimeStat"), zap.Bool("result", false))
				return
			}

			logger.Info("service exit", zap.String("service", "runtimeStat"), zap.Bool("result", true))
		},
	)

	select {
	case <-ctx.Done():
		logger.Warn("receive stop signal....")
		logger.Warn("service shutdown", zap.String("service", "runtimeStat"))
		// 不再接收新请求，等待5秒如果还没有完成，直接强制关闭这些请求
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 不再接收http请求
			if err := rs.srv.Shutdown(ctx); err != nil {
				logger.Error("service shutdown", zap.String("service", "runtimeStat"), zap.Bool("result", false), zap.Error(err))
				return
			}

			logger.Info("service shutdown", zap.String("service", "runtimeStat"), zap.Bool("result", true))
			return
		}()
	}

	return nil
}
