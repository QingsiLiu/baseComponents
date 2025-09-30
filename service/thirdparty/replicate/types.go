package replicate

// PredictionRequest 预测任务请求
type PredictionRequest struct {
	Version string      `json:"version,omitempty"`
	Input   interface{} `json:"input"`
	Webhook string      `json:"webhook,omitempty"`
}

// PredictionResponse 预测任务响应
type PredictionResponse struct {
	ID          string      `json:"id"`
	Model       string      `json:"model"`
	Version     string      `json:"version"`
	Input       interface{} `json:"input"`
	Logs        string      `json:"logs"`
	Output      interface{} `json:"output"`
	DataRemoved bool        `json:"data_removed"`
	Error       interface{} `json:"error"`
	Status      string      `json:"status"`
	CreatedAt   string      `json:"created_at"`
	StartedAt   string      `json:"started_at,omitempty"`
	CompletedAt string      `json:"completed_at,omitempty"`
	URLs        struct {
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
	Metrics struct {
		PredictTime float64 `json:"predict_time,omitempty"`
	} `json:"metrics,omitempty"`
}

// ErrorResponse API错误响应
type ErrorResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// ConvertOutputToStringSlice 将interface{}类型的output转换为字符串切片
func ConvertOutputToStringSlice(output interface{}) []string {
	if output == nil {
		return nil
	}

	switch v := output.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			}
		}
		return result
	case string:
		if v != "" {
			return []string{v}
		}
		return nil
	default:
		return nil
	}
}
