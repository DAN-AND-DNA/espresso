package modules

import (
	"context"
	"espresso/pkg/modules/internal"
)

type Module = internal.Module

func Register(module Module) {
	internal.GetModules().Register(module)
}

func Init(ctx context.Context) {
	internal.GetModules().ModulesInit(ctx)
}

func Exit(ctx context.Context) {
	internal.GetModules().ModulesExit(ctx)
	internal.GetModules().ModulesClean(ctx)
}

// GetModule 获得模块，未安装返回 nil
func GetModule(name string) Module {
	return internal.GetModules().GetModule(name)
}
