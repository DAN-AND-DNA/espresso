// 本文件由工具生成，请勿修改

package metric

import (
	"espresso/modules/metric/internal"
	"espresso/pkg/modules"
)

func init() {
	modules.Register(internal.GetSingleInst())
}
