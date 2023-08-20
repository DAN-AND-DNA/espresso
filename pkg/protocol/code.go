package protocol

const (
	CodeOk = iota

	CodeInternalError    = 10000 + iota // 非法json参数
	CodeInvalidRequest                  // 非法请求
	CodeInvalidJsonParam                // 非法json参数

)
