package internal

import (
	"context"
	_ "espresso/modules" // 注册全部模块到mvc
	"espresso/pkg/ctxhelper"
	"espresso/pkg/gosafe"
	"espresso/pkg/mvc"
	"espresso/pkg/network"
	"espresso/pkg/runtimestat"
	"github.com/dan-and-dna/minilog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"os/signal"
	"sync"
	"syscall"
)

type Plugin = mvc.Plugin
type Handler = mvc.Handler

type options struct {
	logger             *minilog.MiniLog
	registry           *prometheus.Registry
	httpAddress        string
	runtimeStatAddress string
	enablePProf        bool
	app                *App
	injects            map[string]any
	moduleContext      []func(ctx context.Context) context.Context
	modulePlugins      []Plugin
}

type Option func(opts *options)

func Logger(logger *minilog.MiniLog) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func Metric(registry *prometheus.Registry) Option {
	return func(opts *options) {
		opts.registry = registry
	}
}

func Http(httpAddress string) Option {
	return func(opts *options) {
		opts.httpAddress = httpAddress
	}
}

func Grpc(grpcAddress string) Option {
	return func(opts *options) {
		// TODO
	}
}

func RuntimeStat(address string) Option {
	return func(opts *options) {
		opts.runtimeStatAddress = address
	}
}

func PProf(enable bool) Option {
	return func(opts *options) {
		opts.enablePProf = enable
	}
}

func ModuleContext(f func(ctx context.Context) context.Context) Option {
	return func(opts *options) {
		opts.moduleContext = append(opts.moduleContext, f)
	}
}

func ModulePlugin(plugin Plugin) Option {
	return func(opts *options) {
		opts.modulePlugins = append(opts.modulePlugins, plugin)
	}
}

type App struct {
	opts *options

	network     *network.Network
	runtimeStat *runtimestat.RuntimeStat
	registry    *prometheus.Registry
	logger      *minilog.MiniLog
	ctx         context.Context
}

func New(applyOptions ...Option) *App {
	app := &App{
		ctx: context.Background(),
	}
	opts := &options{
		app: app,
	}
	app.opts = opts

	app.logger = minilog.New()

	// 应用配置
	for _, applyOption := range applyOptions {
		applyOption(app.opts)
	}

	// 注入模块context
	for _, moduleContext := range app.opts.moduleContext {
		app.ctx = moduleContext(app.ctx)
	}

	// 日志
	if app.opts.logger != nil {
		app.logger = app.opts.logger
		app.ctx = ctxhelper.InjectLogger(app.ctx, app.logger) // 注入给模块

		app.logger.Info("set app option", zap.Bool("enable", true), zap.String("option", "logger"))
	} else {
		app.logger.Info("set app option", zap.Bool("enable", false), zap.String("option", "logger"))
	}

	// metric
	if app.opts.registry != nil {
		app.registry = app.opts.registry
		app.ctx = ctxhelper.InjectRegistry(app.ctx, app.registry) // 注入给模块

		app.logger.Info("set app option", zap.Bool("enable", true), zap.String("option", "metric"))
	} else {
		app.logger.Info("set app option", zap.Bool("enable", false), zap.String("option", "metric"))
	}

	// 设置mvc插件
	mvc.SetPlugins(opts.modulePlugins...)

	// 网络
	if app.opts.httpAddress != "" {
		app.network = network.New(
			network.PProf(app.opts.enablePProf),               // pprof
			network.Logger(app.logger),                        // 日志
			network.Http(app.opts.httpAddress, mvc.NewHttp()), // 启动http服务
			network.Metric(app.registry),                      // metric
			network.ModuleContext(app.opts.moduleContext...),
		)
		if app.opts.enablePProf {
			app.logger.Info("set app option", zap.Bool("enable", true), zap.String("option", "pprof"))
		} else {
			app.logger.Info("set app option", zap.Bool("enable", false), zap.String("option", "pprof"))
		}
		app.logger.Info("set app option", zap.Bool("enable", true), zap.String("option", "network"))
	} else {
		app.logger.Info("set app option", zap.Bool("enable", false), zap.String("option", "network"))
	}

	// 运行时统计
	if app.opts.runtimeStatAddress != "" {
		app.runtimeStat = runtimestat.New(
			runtimestat.Address(app.opts.runtimeStatAddress),
		)
		app.logger.Info("set app option", zap.Bool("enable", true), zap.String("option", "runtime stat"))
	} else {
		app.logger.Info("set app option", zap.Bool("enable", false), zap.String("option", "runtime stat"))
	}

	return app
}

func (app *App) Run() error {
	if app == nil || app.opts == nil {
		return nil
	}
	defer app.logger.Close()

	// 监听关闭信号
	ctx := app.ctx
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wait sync.WaitGroup

	// 启动网络
	if app.network != nil {
		wait.Add(1)
		gosafe.GoSafe(ctx,
			func(ctx context.Context) {
				// 运行mvc
				mvc.Run(ctx)
				// 运行网络
				if err := app.network.Run(ctx); err != nil {
					app.logger.Error("run network fail", zap.Error(err))
					return
				}
			},
			func(ctx context.Context, err error) {
				defer wait.Done()
				if err != nil {
					app.logger.Error("stop network fail", zap.Error(err))
					return
				}

				app.logger.Info("stop network success")
			},
			func(ctx context.Context) {
				// 暂停mvc
				mvc.Stop(ctx)
			})
	}

	// 启动运行时统计
	if app.runtimeStat != nil {
		wait.Add(1)
		gosafe.GoSafe(ctx,
			func(ctx context.Context) {
				// 运行时统计
				if err := app.runtimeStat.Run(ctx); err != nil {
					app.logger.Info("run runtime stat fail", zap.Error(err))
					return
				}
			},
			func(ctx context.Context, err error) {
				defer wait.Done()
				if err != nil {
					app.logger.Error("stop runtime stat fail", zap.Error(err))
					return
				}

				app.logger.Info("stop runtime stat success")
			},
		)

	}

	wait.Wait()
	<-ctx.Done()
	app.logger.Warn("receive stop signal....")
	stop()
	app.logger.Info("bye~ :)")

	return nil
}
