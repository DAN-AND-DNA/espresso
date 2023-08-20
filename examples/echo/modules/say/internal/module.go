// 本文件由工具生产，请完成模块实现

package internal

import (
	"context"
	"espresso/pkg/modules"
	"espresso/pkg/mvc"
	"sync"
)

var (
	metric *Module = nil
	once   sync.Once
)

// Say 接口定义
type Say interface {
	modules.Module // 模块接口

	// 对外接口
}

func GetSingleInst() Say {
	if metric == nil {
		once.Do(func() {
			metric = new(Module)
		})
	}

	return metric
}

// Module 模块
type Module struct {
}

// ModuleUID 模块uid
func (module *Module) ModuleUID() string {
	return "say_v0.1.0"
}

// ModuleInit 模块初始化
func (module *Module) ModuleInit(ctx context.Context) error {
	// 注册消息处理函数
	mvc.Register("say", "hello", module.Hello)
	mvc.Register("say", "tryPanic", module.TryPanic)
	mvc.Register("say", "slow", module.Slow)

	return nil
}

// ModuleExit 模块退出
func (module *Module) ModuleExit(ctx context.Context) error {
	return nil
}

// ModuleClean 模块清理
func (module *Module) ModuleClean(ctx context.Context) error {
	return nil
}
