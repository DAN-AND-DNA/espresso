package internal

import (
	"context"
	"crypto/sha256"
	"espresso/pkg/ctxhelper"
	"espresso/pkg/events"
	"espresso/pkg/gosafe"
	"espresso/pkg/protocol"
	"fmt"
	ginprom "github.com/dan-and-dna/gin-prom"
	"github.com/dan-and-dna/minilog"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kamilsk/tracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Network struct {
	opts *options

	httpRoute *gin.Engine
	httpSrv   *http.Server
}

type options struct {
	httpAddress    string
	enablePProf    bool
	logger         *minilog.MiniLog
	registry       *prometheus.Registry
	mvcHttpPlugin  gin.HandlerFunc
	injects        []gin.HandlerFunc
	moduleContexts []func(context.Context) context.Context
}

type Option func(options *options)

func Http(address string, mvcHttpPlugin gin.HandlerFunc) Option {
	return func(opts *options) {
		opts.httpAddress = address
		opts.mvcHttpPlugin = mvcHttpPlugin
	}
}

func PProf(enable bool) Option {
	return func(opts *options) {
		opts.enablePProf = enable
	}
}

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

func ModuleContext(f ...func(ctx context.Context) context.Context) Option {
	return func(opts *options) {
		opts.moduleContexts = append(opts.moduleContexts, f...)
	}
}

func New(applyOptions ...Option) *Network {
	gin.SetMode(gin.ReleaseMode)

	network := &Network{
		opts: &options{},
	}

	// 应用全部配置
	for _, applyOption := range applyOptions {
		applyOption(network.opts)
	}

	// http监听端口
	if network.opts.httpAddress != "" {
		network.httpRoute = gin.New()

		network.httpSrv = &http.Server{
			Addr:    network.opts.httpAddress,
			Handler: network.httpRoute,
		}

		// pprof
		if network.opts.enablePProf {
			pprof.Register(network.httpRoute)
		}
	}

	// 路由
	var count atomic.Uint64
	var plugins []gin.HandlerFunc
	if network.opts.registry != nil {
		// metric插件
		metrics := ginprom.NewMetrics("default", network.opts.registry)
		network.httpRoute.GET("/metric", GinPromHandler(promhttp.HandlerFor(network.opts.registry, promhttp.HandlerOpts{})))

		// 其他插件
		network.httpRoute.Use(GinSetRecover())
		if network.opts.mvcHttpPlugin != nil {
			plugins = append(plugins,
				GinSetLogger(network.opts.logger),
				ginprom.Export(metrics),
				GinWitheRequestId(&count),
				GinFetchMetadata(),
				func(c *gin.Context) {
					for _, moduleContext := range network.opts.moduleContexts {
						moduleContext(c)
					}
					c.Next()
				},
			)

			plugins = append(plugins, network.opts.mvcHttpPlugin)

			network.httpRoute.GET("/daydream/:module/:message", plugins...)
			network.httpRoute.POST("/daydream/:module/:message", plugins...)
		}

	} else {
		// 其他插件
		network.httpRoute.Use(GinSetRecover())
		if network.opts.mvcHttpPlugin != nil {
			plugins = append(plugins,
				GinSetLogger(network.opts.logger),
				GinWitheRequestId(&count),
				GinFetchMetadata(),
				func(c *gin.Context) {
					for _, moduleContext := range network.opts.moduleContexts {
						moduleContext(c)
					}
					c.Next()
				},
			)

			plugins = append(plugins, network.opts.mvcHttpPlugin)

			network.httpRoute.GET("/daydream/:module/:message", plugins...)
			network.httpRoute.POST("/daydream/:module/:message", plugins...)
		}
	}

	return network
}

func (network *Network) Run(ctx context.Context) error {
	logger := network.opts.logger
	if err := network.httpRoute.SetTrustedProxies(nil); err != nil {
		logger.Error("service run", zap.Error(err), zap.String("service", "network"), zap.Bool("result", false))
		return err
	}

	// 监听http请求
	gosafe.GoSafe(ctx,
		// 协程逻辑
		func(ctx context.Context) {
			logger.Info("service run", zap.String("listenAddress", network.opts.httpAddress), zap.Bool("result", true), zap.String("service", "network"))
			if err := network.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("service run", zap.Error(err), zap.Bool("result", false), zap.String("service", "network"))
				return
			}
		},
		// 退出前钩子
		func(ctx context.Context, err error) {
			if err != nil {
				logger.Error("service exit", zap.Error(err), zap.String("service", "network"), zap.Bool("result", false))
				return
			}

			logger.Info("service exit", zap.String("service", "network"), zap.Bool("result", true))
		})

	select {
	case <-ctx.Done():
		logger.Warn("receive stop signal....")
		logger.Warn("service shutdown", zap.String("service", "network"))
		// 不再接收新请求，等待5秒如果还没有完成，直接强制关闭这些请求
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 不再接收http请求
			if err := network.httpSrv.Shutdown(ctx); err != nil {
				logger.Error("service shutdown", zap.Error(err), zap.String("service", "network"), zap.Bool("result", false))
				return
			}

			logger.Info("service shutdown", zap.String("service", "network"), zap.Bool("result", true))
		}()
	}

	return nil
}

func GinPromHandler(handler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func GinSetLogger(logger *minilog.MiniLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxhelper.InjectLogger(c, logger)
		c.Next()
	}
}

func GinSetRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			r := recover()
			if r != nil {
				module := c.Param("module")
				message := c.Param("message")
				requestId, _ := ctxhelper.Get(c, "requestId").(string)
				err := fmt.Errorf("%v", r)
				events.PublishModuleCallPanic(c, module, message, err)
				c.JSON(500, protocol.BaseResponse{
					Code:      protocol.CodeInternalError,
					Msg:       "internal error",
					RequestId: requestId,
				})
			}
		}()

		c.Next()
	}
}

func GinWitheRequestId(count *atomic.Uint64) gin.HandlerFunc {
	return func(c *gin.Context) {
		count.Add(1)
		count.CompareAndSwap(math.MaxUint64, 0)

		// FIXME 不是很高效
		paths := []string{c.FullPath(), c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), strconv.FormatUint(count.Load(), 10)}
		requestId := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(paths, "_"))))
		// 注入请求id
		ctxhelper.InjectRequestId(c, requestId)

		// FIXME 精简这种操作
		ctx := tracer.Inject(context.Background(), make([]*tracer.Call, 0, 30))
		t := tracer.Fetch(ctx)

		ctxhelper.InjectTrace(c, t)

		module := c.Param("module")
		message := c.Param("message")
		defer events.PublishModuleCall(c, module, message, time.Now().UnixNano())

		call := t.Start(requestId)
		defer call.Stop()
		c.Next()
	}
}

func GinFetchMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		metadata := make(map[string][]string)

		for key, val := range c.Request.Header {
			metadata[key] = val
		}

		metadata["FullPath"] = []string{c.FullPath()}
		metadata["Uri"] = []string{c.Request.RequestURI}
		metadata["Url"] = []string{c.Request.URL.String()}
		metadata["Module"] = []string{c.Param("module")}
		metadata["Message"] = []string{c.Param("message")}

		ctxhelper.InjectMetadata(c, metadata)

		c.Next()
	}
}
