package modelslab

import (
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"github.com/QingsiLiu/baseComponents/service/text2image"
)

func NewFluxServiceWithClient(client *Client) text2image.Text2ImageService {
	return &FluxService{client: client}
}

func NewInteriorServiceWithClient(client *Client) image2image.Image2ImageService {
	return &InteriorService{client: client}
}

func NewExteriorServiceWithClient(client *Client) image2image.Image2ImageService {
	return &ExteriorService{client: client}
}
