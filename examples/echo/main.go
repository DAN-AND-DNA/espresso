package main

import (
	"context"
	"espresso"
	_ "espresso/examples/echo/modules"
	"espresso/examples/echo/pkg/ctxhelper"
	"espresso/pkg/ctxhelper"
	"espresso/pkg/mvc"
	"github.com/dan-and-dna/minilog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	// 日志
	logger := minilog.New(minilog.Config{
		Environment: "development",
	})

	registry := prometheus.NewRegistry()

	app := espresso.New(
		// 启动日志
		espresso.Logger(logger),
		// 启动metric
		espresso.Metric(registry),
		// 打开pprof
		espresso.PProf(true),
		// 提供http服务
		espresso.Http(":8080"),
		// 提供grpc服务
		espresso.Grpc(":8081"),
		// 运行时统计
		espresso.RuntimeStat(":8079"),
		// 模块context
		espresso.ModuleContext(func(ctx context.Context) context.Context {
			// 注入context
			ctx = mctxhelper.InjectConfigs(ctx, "zzzzzzzzzzz")
			return ctx
		}),
		// 模块调用插件
		espresso.ModulePlugin(plugin("d")),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

type A struct {
	StarTime int64
	EndTime  int64
}

func (a *A) Start(ctx context.Context) func(ctx context.Context) {
	a.StarTime = time.Now().UnixNano()
	return a.End
}

func (a *A) End(ctx context.Context) {
	a.EndTime = time.Now().UnixNano()
	logger := ctxhelper.FetchLogger(ctx)
	logger.Info("call cost", zap.Int64("cost", a.EndTime-a.StarTime))
}

func plugin(newMessage string) mvc.Plugin {

	return func(next mvc.Handler) mvc.Handler {
		a := &A{}
		log.Printf("%p\n", a)
		return func(ctx context.Context, request any, response any) error {
			defer a.Start(ctx)(ctx)

			log.Println("start")
			defer log.Println("end")
			return next(ctx, request, response)
		}
	}
}
