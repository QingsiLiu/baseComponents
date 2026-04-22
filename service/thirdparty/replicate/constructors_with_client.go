package replicate

import (
	"github.com/QingsiLiu/baseComponents/service/aivideo"
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"github.com/QingsiLiu/baseComponents/service/text2image"
)

func NewQwenImageServiceWithClient(client *Client) text2image.Text2ImageService {
	return &QwenImageService{client: client}
}

func NewPrunaAIQwenImageFastServiceWithClient(client *Client) text2image.Text2ImageService {
	return &PrunaAIQwenImageFastService{client: client}
}

func NewFluxSchnellServiceWithClient(client *Client) text2image.Text2ImageService {
	return &FluxSchnellService{client: client}
}

func NewFlux1DevServiceWithClient(client *Client) text2image.Text2ImageService {
	return &Flux1DevService{client: client}
}

func NewNanoBananaServiceWithClient(client *Client) image2image.Image2ImageService {
	return &NanoBananaService{client: client}
}

func NewControlNetServiceWithClient(client *Client) image2image.Image2ImageService {
	return &ControlNetService{client: client}
}

func NewPixverseV5ServiceWithClient(client *Client) aivideo.AIVideoService {
	return &PixverseV5Service{client: client}
}
