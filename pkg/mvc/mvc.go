package mvc

import (
	"context"
	"espresso/pkg/mvc/internal"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type Plugin = internal.Plugin
type Handler = internal.Handler

func NewHttp() gin.HandlerFunc {
	return internal.GetSingleInst().NewHttp()
}

func Run(ctx context.Context) {
	internal.GetSingleInst().Run(ctx)
}

func Stop(ctx context.Context) {
	internal.GetSingleInst().Stop(ctx)
}

func Register(module, message string, handler any) {
	internal.GetSingleInst().Register(module, message, handler)
}

func Publish(event string, args ...any) {
	internal.GetSingleInst().Publish(event, args...)
}

func Subscribe(event string, callback any) error {
	return internal.GetSingleInst().Subscribe(event, callback)
}

func UnSubscribe(event string, callback any) error {
	return internal.GetSingleInst().UnSubscribe(event, callback)
}

func OnTimeDo(spec string, cmd func()) (cron.EntryID, error) {
	return internal.GetSingleInst().OnTimeDo(spec, cmd)
}

func StopOnTimeDo(id cron.EntryID) {
	internal.GetSingleInst().StopOnTimeDo(id)
}

func SetPlugins(plugins ...Plugin) {
	internal.GetSingleInst().SetPlugins(plugins...)
}
