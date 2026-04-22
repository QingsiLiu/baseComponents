package kie

import (
	"github.com/QingsiLiu/baseComponents/service/aivideo"
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"github.com/QingsiLiu/baseComponents/service/text2image"
)

func NewGPTImage2Text2ImageServiceWithClient(client *Client) text2image.Text2ImageService {
	return &GPTImage2Text2ImageService{client: client}
}

func NewQwenText2ImageServiceWithClient(client *Client) text2image.Text2ImageService {
	return &QwenText2ImageService{client: client}
}

func NewIdeogramV3Text2ImageServiceWithClient(client *Client) text2image.Text2ImageService {
	return &IdeogramV3Text2ImageService{client: client}
}

func NewNanoBananaServiceWithClient(client *Client) image2image.Image2ImageService {
	return &NanoBananaService{client: client}
}

func NewKling30VideoServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Kling30VideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling30Video,
			kling30VideoModelName,
			"Kling 3.0 Video",
			client,
			buildKling30VideoInput,
		),
	}
}

func NewKling26ImageToVideoServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Kling26ImageToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26ImageToVideo,
			kling26ImageToVideoModelName,
			"Kling 2.6 Image To Video",
			client,
			buildKling26ImageToVideoInput,
		),
	}
}

func NewKling26TextToVideoServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Kling26TextToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26TextToVideo,
			kling26TextToVideoModelName,
			"Kling 2.6 Text To Video",
			client,
			buildKling26TextToVideoInput,
		),
	}
}

func NewSeedance15ProServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Seedance15ProService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance15Pro,
			seedance15ProModelName,
			"Seedance 1.5 Pro",
			client,
			buildSeedance15ProInput,
		),
	}
}

func NewSeedance2ServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Seedance2Service{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2,
			seedance2ModelName,
			"Seedance 2",
			client,
			buildSeedance2Input,
		),
	}
}

func NewSeedance2FastServiceWithClient(client *Client) aivideo.AIVideoService {
	return &Seedance2FastService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2Fast,
			seedance2FastModelName,
			"Seedance 2 Fast",
			client,
			buildSeedance2Input,
		),
	}
}
