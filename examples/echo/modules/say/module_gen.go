// 本文件由工具生成，请勿修改

package say

import (
	"espresso/examples/echo/modules/say/internal"
	"espresso/pkg/modules"
)

func init() {
	modules.Register(internal.GetSingleInst())
}
