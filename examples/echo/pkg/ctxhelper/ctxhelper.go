package mctxhelper

import (
	"context"
	"espresso/pkg/ctxhelper"
)

func InjectConfigs(ctx context.Context, configs string) context.Context {
	const ConfigsKey = "configs_key"
	ctx = ctxhelper.Set(ctx, ConfigsKey, configs)
	return ctx
}

func FetchConfigs(ctx context.Context) string {
	const ConfigsKey = "configs_key"
	ptr := ctxhelper.Get(ctx, ConfigsKey)
	if ptr == nil {
		return ""
	}

	ret, _ := ptr.(string)
	return ret
}
