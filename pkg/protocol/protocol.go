package protocol

type BaseResponse struct {
	Code      int32  `json:"code"`
	Msg       string `json:"msg,omitempty"`
	RequestId string `json:"requestId,omitempty"`
}
