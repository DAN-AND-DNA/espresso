package internal

import (
	"context"
	"espresso/pkg/ctxhelper"
	"fmt"
	"go.uber.org/zap"
	"sync"
)

var (
	modulesSingleInst *Modules = nil
	once              sync.Once
)

type Modules struct {
	modules       map[string]Module
	moduleUIDs    map[string]struct{}
	modulesOrders []string
	modulesInits  []func(context.Context) error
	modulesExits  []func(context.Context) error
	modulesCleans []func(context.Context) error
}

type Summary struct {
	Module string
	Ok     bool
	Error  error
}

func GetModules() *Modules {
	if modulesSingleInst == nil {
		once.Do(func() {
			modulesSingleInst = new(Modules)
			modulesSingleInst.init()
		})
	}

	return modulesSingleInst
}

// init 模块管理器初始化
func (modules *Modules) init() {
	modulesSingleInst.modules = make(map[string]Module)
}

// Register 注册模块
func (modules *Modules) Register(module Module) {
	uid := module.ModuleUID()
	if uid == "" {
		panic(fmt.Errorf("bad uid: %s", uid))
	}

	if _, ok := modules.moduleUIDs[uid]; ok {
		panic(fmt.Errorf("uid dunplicated: %s", uid))
	}

	initWithPanic := func(ctx context.Context) error {
		defer recoverWithStacktrace(ctx, uid)

		return module.ModuleInit(ctx)
	}

	exitWithRecover := func(ctx context.Context) error {
		defer recoverWithStacktrace(ctx, uid)

		return module.ModuleExit(ctx)
	}

	cleanWithRecover := func(ctx context.Context) error {
		defer recoverWithStacktrace(ctx, uid)

		return module.ModuleClean(ctx)
	}

	modules.modules[uid] = module
	modules.modulesOrders = append(modules.modulesOrders, uid)
	modules.modulesInits = append(modules.modulesInits, initWithPanic)
	modules.modulesExits = append(modules.modulesExits, exitWithRecover)
	modules.modulesCleans = append(modules.modulesCleans, cleanWithRecover)
}

// recoverWithStacktrace 异常处理
func recoverWithStacktrace(ctx context.Context, moduleName string) {
	r := recover()

	if r != nil {
		logger := ctxhelper.FetchLogger(ctx)
		err := fmt.Errorf("%v", r)
		logger.Error("module panic", zap.String("module", moduleName), zap.Error(err))
	}
}

// ModulesInit 初始化全部模块
func (modules *Modules) ModulesInit(ctx context.Context) {
	logger := ctxhelper.FetchLogger(ctx)
	var summaries []Summary

	for index, moduleName := range modules.modulesOrders {
		moduleInit := modules.modulesInits[index]
		if err := moduleInit(ctx); err != nil {
			logger.Error("installing a new module", zap.String("module", moduleName), zap.Bool("result", false), zap.Error(err))
			summaries = append(summaries, Summary{Module: moduleName, Ok: false, Error: err})
			continue
		}

		logger.Info("installing a new module", zap.String("module", moduleName), zap.Bool("result", true))
		summaries = append(summaries, Summary{Module: moduleName, Ok: true})
	}

	// 总结
	successSum := 0
	failSum := 0
	var failModules []string
	for _, summary := range summaries {
		if !summary.Ok {
			failSum++
			failModules = append(failModules, summary.Module)
		} else {
			successSum++
		}
	}

	logger.Info("module installation summary", zap.Int("success", successSum), zap.Int("fail", failSum), zap.Strings("fail modules", failModules))
}

// ModulesClean 模块清理
func (modules *Modules) ModulesClean(ctx context.Context) {
	logger := ctxhelper.FetchLogger(ctx)

	count := len(modules.modulesOrders)
	for index := count - 1; index >= 0; index-- {
		moduleName := modules.modulesOrders[index]
		moduleClean := modules.modulesCleans[index]
		if err := moduleClean(ctx); err != nil {
			logger.Error("cleaning a module", zap.String("module", moduleName), zap.Bool("result", false), zap.Error(err))
			continue
		}

		logger.Info("cleaning a module", zap.String("module", moduleName), zap.Bool("result", true))
	}
}

// ModulesExit 模块退出
func (modules *Modules) ModulesExit(ctx context.Context) {
	logger := ctxhelper.FetchLogger(ctx)

	count := len(modules.modulesOrders)
	for index := count - 1; index >= 0; index-- {
		moduleName := modules.modulesOrders[index]
		modulesExit := modules.modulesExits[index]
		if err := modulesExit(ctx); err != nil {
			logger.Error("uninstalling a module", zap.String("module", moduleName), zap.Bool("result", false), zap.Error(err))
			continue
		}

		logger.Info("uninstalling a module", zap.String("module", moduleName), zap.Bool("result", true))
	}
}

func (modules *Modules) GetModule(name string) Module {
	m, ok := modules.modules[name]
	if !ok {
		panic(fmt.Errorf("no such module:%s", name))
	}

	return m
}
