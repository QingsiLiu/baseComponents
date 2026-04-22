package runtime

import (
	"context"
	"fmt"

	v1aivideo "github.com/QingsiLiu/baseComponents/service/aivideo"
	v1image2image "github.com/QingsiLiu/baseComponents/service/image2image"
	v1text2image "github.com/QingsiLiu/baseComponents/service/text2image"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/kie"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/modelslab"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/replicate"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
	"github.com/QingsiLiu/baseComponents/service/v2/catalog"
	"github.com/QingsiLiu/baseComponents/service/v2/core"
	imageedit "github.com/QingsiLiu/baseComponents/service/v2/image/edit"
	imagegenerate "github.com/QingsiLiu/baseComponents/service/v2/image/generate"
	videogenerate "github.com/QingsiLiu/baseComponents/service/v2/video/generate"
)

type ProviderConfig struct {
	APIKey  string
	BaseURL string
}

type Config struct {
	KIE       ProviderConfig
	WellAPI   ProviderConfig
	Replicate ProviderConfig
	ModelsLab ProviderConfig
}

type Runtime struct {
	catalog              *catalog.Directory
	imageGenerateDrivers map[core.Provider]ImageGenerateDriver
	imageEditDrivers     map[core.Provider]ImageEditDriver
	videoGenerateDrivers map[core.Provider]VideoGenerateDriver
}

type Option func(*Runtime)

func NewBuiltins(cfg Config, opts ...Option) (*Runtime, error) {
	dir, err := catalog.New(catalog.BuiltinOfferings())
	if err != nil {
		return nil, err
	}

	r := &Runtime{
		catalog: dir,
		imageGenerateDrivers: map[core.Provider]ImageGenerateDriver{
			core.ProviderKIE:       &kieImageGenerateDriver{cfg: cfg.KIE},
			core.ProviderWellAPI:   &wellAPIImageGenerateDriver{cfg: cfg.WellAPI},
			core.ProviderReplicate: &replicateImageGenerateDriver{cfg: cfg.Replicate},
			core.ProviderModelsLab: &modelsLabImageGenerateDriver{cfg: cfg.ModelsLab},
		},
		imageEditDrivers: map[core.Provider]ImageEditDriver{
			core.ProviderKIE:       &kieImageEditDriver{cfg: cfg.KIE},
			core.ProviderReplicate: &replicateImageEditDriver{cfg: cfg.Replicate},
			core.ProviderModelsLab: &modelsLabImageEditDriver{cfg: cfg.ModelsLab},
		},
		videoGenerateDrivers: map[core.Provider]VideoGenerateDriver{
			core.ProviderKIE:       &kieVideoGenerateDriver{cfg: cfg.KIE},
			core.ProviderReplicate: &replicateVideoGenerateDriver{cfg: cfg.Replicate},
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func WithImageGenerateDriver(provider core.Provider, driver ImageGenerateDriver) Option {
	return func(r *Runtime) {
		r.imageGenerateDrivers[provider] = driver
	}
}

func WithImageEditDriver(provider core.Provider, driver ImageEditDriver) Option {
	return func(r *Runtime) {
		r.imageEditDrivers[provider] = driver
	}
}

func WithVideoGenerateDriver(provider core.Provider, driver VideoGenerateDriver) Option {
	return func(r *Runtime) {
		r.videoGenerateDrivers[provider] = driver
	}
}

func (r *Runtime) Catalog() *catalog.Directory {
	return r.catalog
}

func (r *Runtime) ImageGenerate() imagegenerate.Service {
	return &imageGenerateService{runtime: r}
}

func (r *Runtime) ImageEdit() imageedit.Service {
	return &imageEditService{runtime: r}
}

func (r *Runtime) VideoGenerate() videogenerate.Service {
	return &videoGenerateService{runtime: r}
}

func (r *Runtime) resolve(target core.Target, capability core.Capability) (catalog.Offering, error) {
	if target.Capability == "" {
		target.Capability = capability
	}
	offering, err := r.catalog.Resolve(target)
	if err != nil {
		return catalog.Offering{}, err
	}
	if offering.Capability != capability {
		return catalog.Offering{}, fmt.Errorf("offering %s has capability %s, want %s", offering.Key, offering.Capability, capability)
	}
	return offering, nil
}

func (r *Runtime) offeringForOperation(key string, capability core.Capability) (catalog.Offering, error) {
	offering, ok := r.catalog.Get(key)
	if !ok {
		return catalog.Offering{}, fmt.Errorf("offering %q not found", key)
	}
	if offering.Capability != capability {
		return catalog.Offering{}, fmt.Errorf("offering %s has capability %s, want %s", offering.Key, offering.Capability, capability)
	}
	return offering, nil
}

type ImageGenerateDriver interface {
	Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error)
	Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error
	Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error
}

type ImageEditDriver interface {
	Run(ctx context.Context, offering catalog.Offering, req *imageedit.Request) (*core.Operation[imageedit.Result], error)
	Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error
	Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error
}

type VideoGenerateDriver interface {
	Run(ctx context.Context, offering catalog.Offering, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error)
	Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error
	Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error
}

type imageGenerateService struct {
	runtime *Runtime
}

func (s *imageGenerateService) Run(ctx context.Context, target core.Target, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	offering, err := s.runtime.resolve(target, core.CapabilityImageGenerate)
	if err != nil {
		return nil, err
	}
	driver, ok := s.runtime.imageGenerateDrivers[offering.Provider]
	if !ok {
		return nil, fmt.Errorf("no image generate driver for provider %s", offering.Provider)
	}
	return driver.Run(ctx, offering, req)
}

func (s *imageGenerateService) Refresh(ctx context.Context, op *core.Operation[imagegenerate.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityImageGenerate)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.imageGenerateDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no image generate driver for provider %s", offering.Provider)
	}
	return driver.Refresh(ctx, offering, op)
}

func (s *imageGenerateService) Cancel(ctx context.Context, op *core.Operation[imagegenerate.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityImageGenerate)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.imageGenerateDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no image generate driver for provider %s", offering.Provider)
	}
	return driver.Cancel(ctx, offering, op)
}

type imageEditService struct {
	runtime *Runtime
}

func (s *imageEditService) Run(ctx context.Context, target core.Target, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	offering, err := s.runtime.resolve(target, core.CapabilityImageEdit)
	if err != nil {
		return nil, err
	}
	driver, ok := s.runtime.imageEditDrivers[offering.Provider]
	if !ok {
		return nil, fmt.Errorf("no image edit driver for provider %s", offering.Provider)
	}
	return driver.Run(ctx, offering, req)
}

func (s *imageEditService) Refresh(ctx context.Context, op *core.Operation[imageedit.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityImageEdit)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.imageEditDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no image edit driver for provider %s", offering.Provider)
	}
	return driver.Refresh(ctx, offering, op)
}

func (s *imageEditService) Cancel(ctx context.Context, op *core.Operation[imageedit.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityImageEdit)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.imageEditDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no image edit driver for provider %s", offering.Provider)
	}
	return driver.Cancel(ctx, offering, op)
}

type videoGenerateService struct {
	runtime *Runtime
}

func (s *videoGenerateService) Run(ctx context.Context, target core.Target, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error) {
	offering, err := s.runtime.resolve(target, core.CapabilityVideoGenerate)
	if err != nil {
		return nil, err
	}
	driver, ok := s.runtime.videoGenerateDrivers[offering.Provider]
	if !ok {
		return nil, fmt.Errorf("no video generate driver for provider %s", offering.Provider)
	}
	return driver.Run(ctx, offering, req)
}

func (s *videoGenerateService) Refresh(ctx context.Context, op *core.Operation[videogenerate.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityVideoGenerate)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.videoGenerateDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no video generate driver for provider %s", offering.Provider)
	}
	return driver.Refresh(ctx, offering, op)
}

func (s *videoGenerateService) Cancel(ctx context.Context, op *core.Operation[videogenerate.Result]) error {
	if op == nil {
		return fmt.Errorf("operation is nil")
	}
	offering, err := s.runtime.offeringForOperation(op.OfferingKey, core.CapabilityVideoGenerate)
	if err != nil {
		return err
	}
	driver, ok := s.runtime.videoGenerateDrivers[offering.Provider]
	if !ok {
		return fmt.Errorf("no video generate driver for provider %s", offering.Provider)
	}
	return driver.Cancel(ctx, offering, op)
}

func text2ImageRequest(req *imagegenerate.Request) *v1text2image.Text2ImageTaskRunReq {
	if req == nil {
		return &v1text2image.Text2ImageTaskRunReq{}
	}
	return &v1text2image.Text2ImageTaskRunReq{
		Prompt:               req.Prompt,
		NegativePrompt:       req.NegativePrompt,
		Seed:                 req.Seed,
		NumOutputs:           req.Count,
		AspectRatio:          req.AspectRatio,
		ImageWidth:           req.Width,
		ImageHeight:          req.Height,
		OutputFormat:         req.OutputFormat,
		OutputQuality:        req.OutputQuality,
		DisableSafetyChecker: req.SafetyMode.IsDisabled(),
	}
}

func image2ImageRequest(req *imageedit.Request) *v1image2image.Image2ImageTaskRunReq {
	if req == nil {
		return &v1image2image.Image2ImageTaskRunReq{}
	}
	return &v1image2image.Image2ImageTaskRunReq{
		Prompt:          req.Prompt,
		NegativePrompt:  req.NegativePrompt,
		ImageInputs:     append([]string(nil), req.Images...),
		Seed:            req.Seed,
		Strength:        req.Strength,
		OutputImageSize: req.AspectRatio,
		OutputFormat:    req.OutputFormat,
		OutputQuality:   req.OutputQuality,
	}
}

func aiVideoRequest(req *videogenerate.Request) *v1aivideo.AIVideoTaskRunReq {
	if req == nil {
		return &v1aivideo.AIVideoTaskRunReq{}
	}
	out := &v1aivideo.AIVideoTaskRunReq{
		Prompt:             req.Prompt,
		ReferenceImageURLs: append([]string(nil), req.ReferenceImages...),
		ReferenceVideoURLs: append([]string(nil), req.ReferenceVideos...),
		Seed:               req.Seed,
		Duration:           req.DurationSeconds,
		AspectRatio:        req.AspectRatio,
		Resolution:         req.Resolution,
		NSFWChecker:        boolPtr(!req.SafetyMode.IsDisabled()),
	}
	if req.GenerateAudio != nil {
		out.GenerateAudio = boolPtr(*req.GenerateAudio)
	}
	if len(req.ReferenceImages) > 0 {
		out.Image = req.ReferenceImages[0]
	}
	return out
}

func boolPtr(v bool) *bool {
	return &v
}

func cloneBoolPtr(v *bool) *bool {
	if v == nil {
		return nil
	}
	return boolPtr(*v)
}

func imageResultFromURLs(urls []string) imagegenerate.Result {
	images := make([]imagegenerate.Image, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			continue
		}
		images = append(images, imagegenerate.Image{URL: url})
	}
	return imagegenerate.Result{Images: images}
}

func imageEditResultFromURLs(urls []string) imageedit.Result {
	images := make([]imageedit.Image, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			continue
		}
		images = append(images, imageedit.Image{URL: url})
	}
	return imageedit.Result{Images: images}
}

func videoResultFromURLs(urls []string) videogenerate.Result {
	videos := make([]videogenerate.Video, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			continue
		}
		videos = append(videos, videogenerate.Video{URL: url})
	}
	return videogenerate.Result{Videos: videos}
}

func textStatus(status int32) core.OperationStatus {
	switch status {
	case v1text2image.TaskStatusCompleted:
		return core.OperationStatusCompleted
	case v1text2image.TaskStatusRunning:
		return core.OperationStatusRunning
	case v1text2image.TaskStatusFailed:
		return core.OperationStatusFailed
	case v1text2image.TaskStatusCanceled:
		return core.OperationStatusCanceled
	default:
		return core.OperationStatusPending
	}
}

func imageStatus(status int32) core.OperationStatus {
	switch status {
	case v1image2image.TaskStatusCompleted:
		return core.OperationStatusCompleted
	case v1image2image.TaskStatusRunning:
		return core.OperationStatusRunning
	case v1image2image.TaskStatusFailed:
		return core.OperationStatusFailed
	case v1image2image.TaskStatusCanceled:
		return core.OperationStatusCanceled
	default:
		return core.OperationStatusPending
	}
}

func videoStatus(status int32) core.OperationStatus {
	switch status {
	case v1aivideo.TaskStatusCompleted:
		return core.OperationStatusCompleted
	case v1aivideo.TaskStatusRunning:
		return core.OperationStatusRunning
	case v1aivideo.TaskStatusFailed:
		return core.OperationStatusFailed
	case v1aivideo.TaskStatusCanceled:
		return core.OperationStatusCanceled
	default:
		return core.OperationStatusPending
	}
}

func newKIEClient(cfg ProviderConfig) *kie.Client {
	return kie.NewClientWithConfig(kie.Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
}

func newReplicateClient(cfg ProviderConfig) *replicate.Client {
	return replicate.NewClientWithConfig(replicate.Config{
		APIToken: cfg.APIKey,
		BaseURL:  cfg.BaseURL,
	})
}

func newModelsLabClient(cfg ProviderConfig) *modelslab.Client {
	return modelslab.NewClientWithConfig(modelslab.Config{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
}

func textServiceForOffering(offering catalog.Offering, cfg ProviderConfig) (v1text2image.Text2ImageService, error) {
	switch offering.Provider {
	case core.ProviderKIE:
		client := newKIEClient(cfg)
		switch offering.Model {
		case core.ModelGPTImage2:
			return kie.NewGPTImage2Text2ImageServiceWithClient(client), nil
		case core.ModelQwenImage:
			return kie.NewQwenText2ImageServiceWithClient(client), nil
		case core.ModelIdeogramV3:
			return kie.NewIdeogramV3Text2ImageServiceWithClient(client), nil
		}
	case core.ProviderReplicate:
		client := newReplicateClient(cfg)
		switch offering.Model {
		case core.ModelQwenImage:
			return replicate.NewQwenImageServiceWithClient(client), nil
		case core.ModelQwenImageFast:
			return replicate.NewPrunaAIQwenImageFastServiceWithClient(client), nil
		case core.ModelFluxSchnell:
			return replicate.NewFluxSchnellServiceWithClient(client), nil
		case core.ModelFlux1Dev:
			return replicate.NewFlux1DevServiceWithClient(client), nil
		}
	case core.ProviderModelsLab:
		client := newModelsLabClient(cfg)
		if offering.Model == core.ModelModelsLabFlux {
			return modelslab.NewFluxServiceWithClient(client), nil
		}
	}
	return nil, fmt.Errorf("no text image service for offering %s", offering.Key)
}

func imageServiceForOffering(offering catalog.Offering, cfg ProviderConfig) (v1image2image.Image2ImageService, error) {
	switch offering.Provider {
	case core.ProviderKIE:
		client := newKIEClient(cfg)
		if offering.Model == core.ModelNanoBanana {
			return kie.NewNanoBananaServiceWithClient(client), nil
		}
	case core.ProviderReplicate:
		client := newReplicateClient(cfg)
		switch offering.Model {
		case core.ModelNanoBanana:
			return replicate.NewNanoBananaServiceWithClient(client), nil
		case core.ModelControlNet:
			return replicate.NewControlNetServiceWithClient(client), nil
		}
	case core.ProviderModelsLab:
		client := newModelsLabClient(cfg)
		switch offering.Model {
		case core.ModelModelsLabInterior:
			return modelslab.NewInteriorServiceWithClient(client), nil
		case core.ModelModelsLabExterior:
			return modelslab.NewExteriorServiceWithClient(client), nil
		}
	}
	return nil, fmt.Errorf("no image edit service for offering %s", offering.Key)
}

func videoServiceForOffering(offering catalog.Offering, cfg ProviderConfig) (v1aivideo.AIVideoService, error) {
	switch offering.Provider {
	case core.ProviderKIE:
		client := newKIEClient(cfg)
		switch offering.Model {
		case core.ModelKling26ImageToVideo:
			return kie.NewKling26ImageToVideoServiceWithClient(client), nil
		case core.ModelKling26TextToVideo:
			return kie.NewKling26TextToVideoServiceWithClient(client), nil
		case core.ModelKling30Video:
			return kie.NewKling30VideoServiceWithClient(client), nil
		case core.ModelSeedance15Pro:
			return kie.NewSeedance15ProServiceWithClient(client), nil
		case core.ModelSeedance2:
			return kie.NewSeedance2ServiceWithClient(client), nil
		case core.ModelSeedance2Fast:
			return kie.NewSeedance2FastServiceWithClient(client), nil
		}
	case core.ProviderReplicate:
		client := newReplicateClient(cfg)
		if offering.Model == core.ModelPixverseV5 {
			return replicate.NewPixverseV5ServiceWithClient(client), nil
		}
	}
	return nil, fmt.Errorf("no video service for offering %s", offering.Key)
}

type asyncImageGenerateDriver struct {
	cfg ProviderConfig
}

func (d *asyncImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	service, err := textServiceForOffering(offering, d.cfg)
	if err != nil {
		return nil, err
	}
	taskID, err := service.TaskRun(text2ImageRequest(req))
	if err != nil {
		return nil, err
	}
	return &core.Operation[imagegenerate.Result]{
		OfferingKey: offering.Key,
		ExternalID:  taskID,
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusPending,
	}, nil
}

func (d *asyncImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	service, err := textServiceForOffering(offering, d.cfg)
	if err != nil {
		return err
	}
	task, err := service.TaskGet(op.ExternalID)
	if err != nil {
		return err
	}
	op.Status = textStatus(task.Status)
	op.Result = imageResultFromURLs(task.Result)
	return nil
}

func (d *asyncImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	service, err := textServiceForOffering(offering, d.cfg)
	if err != nil {
		return err
	}
	if err := service.TaskCancel(op.ExternalID); err != nil {
		return err
	}
	op.Status = core.OperationStatusCanceled
	return nil
}

type kieImageGenerateDriver struct{ cfg ProviderConfig }

func (d *kieImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Run(ctx, offering, req)
}

func (d *kieImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Refresh(ctx, offering, op)
}

func (d *kieImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Cancel(ctx, offering, op)
}

type replicateImageGenerateDriver struct{ cfg ProviderConfig }

type modelsLabImageGenerateDriver struct{ cfg ProviderConfig }

func (d *replicateImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Run(ctx, offering, req)
}

func (d *replicateImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Refresh(ctx, offering, op)
}

func (d *replicateImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Cancel(ctx, offering, op)
}

func (d *modelsLabImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Run(ctx, offering, req)
}

func (d *modelsLabImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Refresh(ctx, offering, op)
}

func (d *modelsLabImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return (&asyncImageGenerateDriver{cfg: d.cfg}).Cancel(ctx, offering, op)
}

type wellAPIImageGenerateDriver struct {
	cfg ProviderConfig
}

func (d *wellAPIImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	if req == nil {
		req = &imagegenerate.Request{}
	}
	client := wellapi.NewClientWithConfig(wellapi.Config{APIKey: d.cfg.APIKey, BaseURL: d.cfg.BaseURL})
	resp, err := client.CreateImageGeneration(&wellapi.ImageGenerateReq{
		Model:          wellapi.ModelGPTImage2,
		Prompt:         req.Prompt,
		N:              req.Count,
		Size:           imageSize(req),
		ResponseFormat: "url",
	})
	if err != nil {
		return nil, err
	}
	images := make([]imagegenerate.Image, 0, len(resp.Data))
	for _, image := range resp.Data {
		images = append(images, imagegenerate.Image{
			URL:           image.URL,
			B64JSON:       image.B64JSON,
			RevisedPrompt: image.RevisedPrompt,
		})
	}
	return &core.Operation[imagegenerate.Result]{
		OfferingKey: offering.Key,
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusCompleted,
		Result:      imagegenerate.Result{Images: images},
	}, nil
}

func (d *wellAPIImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return nil
}

func (d *wellAPIImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	return core.ErrUnsupported
}

func imageSize(req *imagegenerate.Request) string {
	if req != nil && req.Width > 0 && req.Height > 0 {
		return fmt.Sprintf("%dx%d", req.Width, req.Height)
	}
	return "auto"
}

type kieImageEditDriver struct{ cfg ProviderConfig }
type replicateImageEditDriver struct{ cfg ProviderConfig }
type modelsLabImageEditDriver struct{ cfg ProviderConfig }

func (d *kieImageEditDriver) Run(ctx context.Context, offering catalog.Offering, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	return runImageEdit(ctx, offering, d.cfg, req)
}
func (d *kieImageEditDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return refreshImageEdit(ctx, offering, d.cfg, op)
}
func (d *kieImageEditDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return cancelImageEdit(ctx, offering, d.cfg, op)
}

func (d *replicateImageEditDriver) Run(ctx context.Context, offering catalog.Offering, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	return runImageEdit(ctx, offering, d.cfg, req)
}
func (d *replicateImageEditDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return refreshImageEdit(ctx, offering, d.cfg, op)
}
func (d *replicateImageEditDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return cancelImageEdit(ctx, offering, d.cfg, op)
}

func (d *modelsLabImageEditDriver) Run(ctx context.Context, offering catalog.Offering, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	return runImageEdit(ctx, offering, d.cfg, req)
}
func (d *modelsLabImageEditDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return refreshImageEdit(ctx, offering, d.cfg, op)
}
func (d *modelsLabImageEditDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	return cancelImageEdit(ctx, offering, d.cfg, op)
}

func runImageEdit(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	service, err := imageServiceForOffering(offering, cfg)
	if err != nil {
		return nil, err
	}
	taskID, err := service.TaskRun(image2ImageRequest(req))
	if err != nil {
		return nil, err
	}
	return &core.Operation[imageedit.Result]{
		OfferingKey: offering.Key,
		ExternalID:  taskID,
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusPending,
	}, nil
}

func refreshImageEdit(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, op *core.Operation[imageedit.Result]) error {
	service, err := imageServiceForOffering(offering, cfg)
	if err != nil {
		return err
	}
	task, err := service.TaskGet(op.ExternalID)
	if err != nil {
		return err
	}
	op.Status = imageStatus(task.Status)
	op.Result = imageEditResultFromURLs(task.Result)
	return nil
}

func cancelImageEdit(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, op *core.Operation[imageedit.Result]) error {
	service, err := imageServiceForOffering(offering, cfg)
	if err != nil {
		return err
	}
	if err := service.TaskCancel(op.ExternalID); err != nil {
		return err
	}
	op.Status = core.OperationStatusCanceled
	return nil
}

type kieVideoGenerateDriver struct{ cfg ProviderConfig }
type replicateVideoGenerateDriver struct{ cfg ProviderConfig }

func (d *kieVideoGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error) {
	return runVideoGenerate(ctx, offering, d.cfg, req)
}
func (d *kieVideoGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	return refreshVideoGenerate(ctx, offering, d.cfg, op)
}
func (d *kieVideoGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	return cancelVideoGenerate(ctx, offering, d.cfg, op)
}

func (d *replicateVideoGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error) {
	return runVideoGenerate(ctx, offering, d.cfg, req)
}
func (d *replicateVideoGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	return refreshVideoGenerate(ctx, offering, d.cfg, op)
}
func (d *replicateVideoGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	return cancelVideoGenerate(ctx, offering, d.cfg, op)
}

func runVideoGenerate(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error) {
	service, err := videoServiceForOffering(offering, cfg)
	if err != nil {
		return nil, err
	}
	taskID, err := service.TaskRun(aiVideoRequest(req))
	if err != nil {
		return nil, err
	}
	return &core.Operation[videogenerate.Result]{
		OfferingKey: offering.Key,
		ExternalID:  taskID,
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusPending,
	}, nil
}

func refreshVideoGenerate(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, op *core.Operation[videogenerate.Result]) error {
	service, err := videoServiceForOffering(offering, cfg)
	if err != nil {
		return err
	}
	task, err := service.TaskGet(op.ExternalID)
	if err != nil {
		return err
	}
	op.Status = videoStatus(task.Status)
	op.Result = videoResultFromURLs(task.Result)
	return nil
}

func cancelVideoGenerate(ctx context.Context, offering catalog.Offering, cfg ProviderConfig, op *core.Operation[videogenerate.Result]) error {
	service, err := videoServiceForOffering(offering, cfg)
	if err != nil {
		return err
	}
	if err := service.TaskCancel(op.ExternalID); err != nil {
		return err
	}
	op.Status = core.OperationStatusCanceled
	return nil
}
