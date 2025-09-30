package modelslab

import "time"

// API基础配置
const (
	BaseURL           = "https://modelslab.com/api/v6"
	Text2ImgEndpoint  = BaseURL + "/images/text2img"
	InteriorEndpoint  = BaseURL + "/interior/make"
	ExteriorEndpoint  = BaseURL + "/interior/exterior_restorer"
	FetchEndpoint     = BaseURL + "/images/fetch"
	DefaultTimeout    = 30 * time.Second
	APIKeyEnvVar      = "MODELSLAB_API_KEY"
)