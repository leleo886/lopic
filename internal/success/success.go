package success

// SuccessResponse 基本成功响应结构体
type SuccessResponse struct {
	Message string `json:"message"`
}

// DataResponse 带数据的成功响应结构体
type DataResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewSuccessResponse 创建基本成功响应
func NewSuccessResponse(message string) *SuccessResponse {
	return &SuccessResponse{
		Message: message,
	}
}

// NewDataResponse 创建带数据的成功响应
func NewDataResponse(message string, data interface{}) *DataResponse {
	return &DataResponse{
		Message: message,
		Data:    data,
	}
}