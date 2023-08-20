package gosafe

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

var ErrorGOSafePanic = errors.New("go safe panic")

// GoSafe 安全协程执行，panic自动清理
func GoSafe(ctx context.Context, fn func(context.Context), beforeExit func(context.Context, error), cleanups ...func(context.Context)) chan<- error {
	panicChan := make(chan<- error, 1)

	go func() {
		defer func() {
			// 先清理
			func() {
				defer func() {
					// 忽略清理错误
					_ = recover()
				}()

				for _, cleanup := range cleanups {
					cleanup(ctx)
				}
			}()

			// 遇到的崩溃
			r := recover()
			var err error
			if r != nil {
				err = fmt.Errorf("%w reason: %v", ErrorGOSafePanic, r)
			}

			// 退出前通知
			beforeExit(ctx, err)

			panicChan <- err
			close(panicChan)
		}()

		if fn != nil {
			fn(ctx)
		}
	}()

	return panicChan
}
