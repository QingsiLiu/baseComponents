# WellAPI

`wellapi` 包当前提供三类能力：

- LLM provider
  - `NewGeminiService`：封装 Gemini 原生 `generateContent`
  - `NewOpenAIService`：封装 OpenAI 兼容的 `/v1/responses` 与 `/v1/chat/completions`
- Image provider
  - `NewImageService`：封装同步 `POST /v1/images/generations`
- Kling 异步任务服务
  - `NewKlingMotionControlService`
  - `NewKlingEffectsService`

当前能力：

- 文本生成
- 图片等 base64 内联输入
- 同步图片生成（`gpt-image-2` 默认模型，可覆盖）
- 结构化 JSON 输出
- 函数调用
- 模型列表查询
- Gemini 原生扩展：URL Context、Google Search、Code Execution
- Kling 异步任务：动作控制、视频特效

设计约束：

- Gemini provider 继续优先封装 **Gemini 原生格式**
- OpenAI provider 在 v1 只支持公共抽象里的通用子集，不暴露 Responses 专属字段
- Image provider 直接暴露 WellAPI 原生同步接口，不实现 `service/text2image` 的任务语义
- OpenAI provider 暂不支持 `EnableURLContext` / `EnableGoogleSearch` / `EnableCodeExecution`
- 调用方负责将媒体内容编码为 base64，本 SDK 不负责下载远程文件
- Gemini 默认会发送 `thinkingBudget=0`，避免空文本或只消耗 reasoning token

已内置的常用模型：

- Gemini：`gemini-3.1-flash-lite-preview`、`gemini-3.1-pro-preview`、`gemini-3-pro-preview`、`gemini-3-flash-preview`、`gemini-2.5-pro`、`gemini-2.5-flash`
- OpenAI：`gpt-5.4-mini`、`gpt-5.4-mini-2026-03-17`、`gpt-5.4-nano`、`gpt-5.4-nano-2026-03-17`、`gpt-5.4`、`gpt-5-mini-2025-08-07`、`gpt-5-nano-2025-08-07`
- Image：`gpt-image-2`

## 环境变量

```bash
export WELLAPI_API_KEY="your-token"
export WELLAPI_BASE_URL="https://wellapi.ai" # 可选
```

## Kling 异步任务

Kling 相关能力属于异步任务协议，不走 `service/llm`，而是走 `service/aivideo/kling` 定义的领域接口：

- `wellapi.NewKlingMotionControlService()`
- `wellapi.NewKlingEffectsService()`

调用方可以直接使用：

```go
package main

import (
	"log"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewKlingMotionControlService()

	taskID, err := service.TaskRun(&aivideokling.KlingMotionControlTaskRunReq{
		ImageURL:             "https://example.com/image.png",
		VideoURL:             "https://example.com/video.mp4",
		CharacterOrientation: "image",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(taskID)
}
```

## Gemini 文本生成

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

## OpenAI 文本生成

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/llm"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewOpenAIService()

	resp, err := service.Generate(&llm.GenerateReq{
		Model: wellapi.ModelGPT54Mini,
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

## OpenAI 图片生成

`wellapi` 的图片生成当前是同步接口，不走 `service/text2image`：

```go
package main

import (
	"log"

	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func main() {
	service := wellapi.NewImageService()

	resp, err := service.Generate(&wellapi.ImageGenerateReq{
		Prompt: "A watercolor otter reading a newspaper in a cafe.",
		Size:   "1024x1536",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, image := range resp.Data {
		log.Println(image.URL, image.RevisedPrompt)
	}
}
```

如果需要 base64 结果，可以显式指定：

```go
resp, err := service.Generate(&wellapi.ImageGenerateReq{
	Model:          wellapi.ModelGPTImage2,
	Prompt:         "A watercolor otter reading a newspaper in a cafe.",
	ResponseFormat: "b64_json",
})
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
