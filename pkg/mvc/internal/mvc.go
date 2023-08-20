package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	eventbus "github.com/asaskevich/EventBus"
	dispatcher "github.com/dan-and-dna/gin-dispatcher"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"

	"espresso/pkg/ctxhelper"
	"espresso/pkg/modules"
	"espresso/pkg/protocol"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	singleInst *Mvc = nil
	once       sync.Once
)

type Plugin = dispatcher.Plugin
type Handler = dispatcher.Handler

type Mvc struct {
	networkDispatcher *dispatcher.Messages
	modules           map[string]map[string]any

	// 事件管理器
	eventDispatcher eventbus.Bus

	// 定时器
	timer *cron.Cron

	// 暂停
	stop  bool
	stopM sync.RWMutex
}

func GetSingleInst() *Mvc {
	if singleInst == nil {
		once.Do(func() {
			singleInst = new(Mvc)
			singleInst.Init()
		})
	}

	return singleInst
}

func (mvc *Mvc) NewHttp() gin.HandlerFunc {
	// 派发消息
	return dispatcher.GinDispatcher(mvc.networkDispatcher)
}

func (mvc *Mvc) Init() {
	messages := dispatcher.NewMessages()

	messages.MessageId = func(c *gin.Context) string {
		module := c.Param("module")
		message := c.Param("message")

		topic := []string{module, message}
		return strings.Join(topic, "::")
	}

	messages.ShouldBind = func(c *gin.Context, in any) error {
		// 依赖gin做参数检查

		if c.Request.Method == "GET" {
			err := c.ShouldBindQuery(in)
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
		} else {
			err := c.ShouldBindJSON(in)
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
		}

		return nil
	}

	messages.HandleError = func(c *gin.Context, err error) {
		requestId, _ := ctxhelper.Get(c, "requestId").(string)

		if _, ok := err.(*json.InvalidUnmarshalError); ok {
			c.JSON(http.StatusOK, protocol.BaseResponse{
				Code:      protocol.CodeInvalidJsonParam,
				Msg:       "非法json对象",
				RequestId: requestId,
			})
			return
		}

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			c.JSON(http.StatusOK, protocol.BaseResponse{
				Code:      protocol.CodeInvalidJsonParam,
				Msg:       "非法json对象",
				RequestId: requestId,
			})
			return
		}

		// 参数校验错误
		if _, ok := err.(*validator.InvalidValidationError); ok {
			c.JSON(http.StatusOK, protocol.BaseResponse{
				Code:      protocol.CodeInvalidRequest,
				Msg:       "非法参数",
				RequestId: requestId,
			})
			return
		}

		// 参数校验错误
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusOK, protocol.BaseResponse{
				Code:      protocol.CodeInvalidRequest,
				Msg:       "非法参数",
				RequestId: requestId,
			})
			return
		}

		c.JSON(http.StatusOK, protocol.BaseResponse{
			Code:      protocol.CodeInvalidRequest,
			Msg:       err.Error(),
			RequestId: requestId,
		})
	}

	// 网络
	mvc.networkDispatcher = messages

	// 事件
	mvc.eventDispatcher = eventbus.New()

	// 定时器
	mvc.timer = cron.New()
}

func (mvc *Mvc) Run(ctx context.Context) {
	// 初始化模块
	modules.Init(ctx)
	// 启动定时器
	mvc.timer.Start()
}

func (mvc *Mvc) Stop(ctx context.Context) {
	// 等待正在运行的定时任务完成
	<-mvc.timer.Stop().Done()

	// 锁住不再能发布消息，避免模块忘了取消订阅消息
	mvc.stopM.Lock()
	defer mvc.stopM.Unlock()
	mvc.stop = true

	// 模块退出
	modules.Exit(ctx)
}

func (mvc *Mvc) Register(module, message string, handler any) {
	if mvc.modules == nil {
		mvc.modules = make(map[string]map[string]any)
	}

	if fs, ok := mvc.modules[module]; ok {
		if _, ok := fs[message]; ok {
			// 已经被注册
			panic(fmt.Sprintf("module: %s message: %s already be registered", module, message))
		}
	} else {
		mvc.modules[module] = make(map[string]any)
	}

	mvc.networkDispatcher.Register(strings.Join([]string{module, message}, "::"), handler)
	mvc.modules[module][message] = struct{}{}
}

func (mvc *Mvc) Publish(event string, args ...any) {
	mvc.stopM.RLock()
	defer mvc.stopM.RUnlock()

	if mvc.stop == true {
		// 退出阶段就不派发消息，免得模块资源释放，消息还是需要处理导致的问题
		return
	}
	mvc.eventDispatcher.Publish(event, args...)
}

func (mvc *Mvc) Subscribe(event string, callback any) error {
	return mvc.eventDispatcher.Subscribe(event, callback)
}

func (mvc *Mvc) UnSubscribe(event string, callback any) error {
	return mvc.eventDispatcher.Unsubscribe(event, callback)
}

func (mvc *Mvc) OnTimeDo(spec string, cmd func()) (cron.EntryID, error) {
	return mvc.timer.AddFunc(spec, cmd)
}

func (mvc *Mvc) StopOnTimeDo(id cron.EntryID) {
	mvc.timer.Remove(id)
}

func (mvc *Mvc) SetPlugins(plugins ...Plugin) {
	if mvc == nil || mvc.networkDispatcher == nil {
		return
	}

	mvc.networkDispatcher.SetPlugins(plugins...)
}
