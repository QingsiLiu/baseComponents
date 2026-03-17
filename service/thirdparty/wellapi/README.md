# WellAPI Gemini

`wellapi` 包提供基于 **WellAPI + Gemini 原生 `generateContent`** 的正式实现，推荐作为多模态 LLM 的统一接入层。

当前能力：

- 文本生成
- 图片 / 文档 / 音频 / 视频等 base64 内联输入
- 结构化 JSON 输出
- 函数调用
- URL Context
- Google Search
- Code Execution
- 模型列表查询

设计约束：

- 正式 API 只封装 **Gemini 原生格式**
- 不推荐业务方直接使用 `/v1/chat/completions`
- 调用方负责将媒体内容编码为 base64，本 SDK 不负责下载远程文件
- 默认会发送 `thinkingBudget=0`，避免空文本或只消耗 reasoning token

## 环境变量

```bash
export WELLAPI_API_KEY="your-token"
export WELLAPI_BASE_URL="https://wellapi.ai" # 可选
```

## 文本生成

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewGeminiService()

	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Reply with OK only."},
				},
			},
		},
		MaxOutputTokens: 16,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.Text)
}
```

## 图片理解（base64 image）

```go
package main

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	imageBytes, err := os.ReadFile("example.png")
	if err != nil {
		log.Fatal(err)
	}

	service := wellapi.NewGeminiService()
	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Describe this image in one sentence."},
					{
						MimeType:         "image/png",
						InlineDataBase64: base64.StdEncoding.EncodeToString(imageBytes),
					},
				},
			},
		},
		MaxOutputTokens: 64,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.Text)
}
```

## 结构化 JSON 输出

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewGeminiService()

	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Return a JSON object with name=assistant and ok=true."},
				},
			},
		},
		ResponseSchema: map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"name": map[string]any{"type": "STRING"},
				"ok":   map[string]any{"type": "BOOLEAN"},
			},
			"required":         []string{"name", "ok"},
			"propertyOrdering": []string{"name", "ok"},
		},
		MaxOutputTokens: 64,
	})
	if err != nil {
		log.Fatal(err)
	}

	var out struct {
		Name string `json:"name"`
		OK   bool   `json:"ok"`
	}
	if err := resp.DecodeJSON(&out); err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", out)
}
```

## 函数调用

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewGeminiService()

	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Schedule a meeting with Alice tomorrow at 10:00 about launch."},
				},
			},
		},
		Tools: []llm.ToolSpec{
			{
				Name:        "schedule_meeting",
				Description: "Schedule a meeting",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"attendees": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"date":  map[string]any{"type": "string"},
						"time":  map[string]any{"type": "string"},
						"topic": map[string]any{"type": "string"},
					},
					"required": []string{"attendees", "date", "time", "topic"},
				},
			},
		},
		MaxOutputTokens: 128,
	})
	if err != nil {
		log.Fatal(err)
	}

	call := resp.FirstFunctionCall()
	if call == nil {
		log.Fatal("no function call returned")
	}

	log.Printf("function=%s args=%v\n", call.Name, call.Args)
}
```

## URL Context

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewGeminiService()

	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Summarize https://ai.google.dev/gemini-api/docs/models in one sentence."},
				},
			},
		},
		EnableURLContext: true,
		MaxOutputTokens: 128,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.Text)
}
```

## 模型列表查询

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	client := wellapi.NewClient()

	resp, err := client.ListModels()
	if err != nil {
		log.Fatal(err)
	}

	for _, model := range resp.Data {
		log.Println(model.ID, model.SupportedEndpointTypes)
	}
}
```
