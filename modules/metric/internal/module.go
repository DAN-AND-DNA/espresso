// 本文件由工具生产，请完成模块实现

package internal

import (
	"context"
	"espresso/pkg/modules"
	"sync"
)

var (
	metric *Module = nil
	once   sync.Once
)

// Metric 接口定义
type Metric interface {
	modules.Module // 模块接口

	// 对外接口
}

func GetSingleInst() Metric {
	if metric == nil {
		once.Do(func() {
			metric = new(Module)
		})
	}

	return metric
}

// Module 模块
type Module struct {
	*Model
	network *Controller
}

// ModuleUID 模块uid
func (module *Module) ModuleUID() string {
	return "metric_v0.1.0_core"
}

// ModuleInit 模块初始化
func (module *Module) ModuleInit(ctx context.Context) error {
	module.Model = new(Model)
	if err := module.Init(ctx); err != nil {
		return err
	}

	module.network = new(Controller)
	if err := module.network.Init(module.Model); err != nil {
		return err
	}

	return nil
}

// ModuleExit 模块退出
func (module *Module) ModuleExit(ctx context.Context) error {
	return nil
}

// ModuleClean 模块清理
func (module *Module) ModuleClean(ctx context.Context) error {
	module.network.Clean()

	return nil
}
