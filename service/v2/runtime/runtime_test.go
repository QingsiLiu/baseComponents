package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

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

type fakeImageGenerateDriver struct {
	mode        core.ExecutionMode
	runStatus   core.OperationStatus
	result      imagegenerate.Result
	refreshHits int
	cancelErr   error
}

type fakeImageEditDriver struct{}

func (d *fakeImageEditDriver) Run(ctx context.Context, offering catalog.Offering, req *imageedit.Request) (*core.Operation[imageedit.Result], error) {
	return &core.Operation[imageedit.Result]{
		OfferingKey: offering.Key,
		ExternalID:  "edit-1",
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusPending,
	}, nil
}

func (d *fakeImageEditDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	op.Status = core.OperationStatusCompleted
	op.Result = imageedit.Result{Images: []imageedit.Image{{URL: "https://example.com/edit.png"}}}
	return nil
}

func (d *fakeImageEditDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imageedit.Result]) error {
	op.Status = core.OperationStatusCanceled
	return nil
}

type fakeVideoGenerateDriver struct{}

func (d *fakeVideoGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *videogenerate.Request) (*core.Operation[videogenerate.Result], error) {
	return &core.Operation[videogenerate.Result]{
		OfferingKey: offering.Key,
		ExternalID:  "video-1",
		Mode:        offering.ExecutionMode,
		Status:      core.OperationStatusPending,
	}, nil
}

func (d *fakeVideoGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	op.Status = core.OperationStatusCompleted
	op.Result = videogenerate.Result{Videos: []videogenerate.Video{{URL: "https://example.com/video.mp4"}}}
	return nil
}

func (d *fakeVideoGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[videogenerate.Result]) error {
	op.Status = core.OperationStatusCanceled
	return nil
}

func (d *fakeImageGenerateDriver) Run(ctx context.Context, offering catalog.Offering, req *imagegenerate.Request) (*core.Operation[imagegenerate.Result], error) {
	return &core.Operation[imagegenerate.Result]{
		OfferingKey: offering.Key,
		ExternalID:  "external-1",
		Mode:        d.mode,
		Status:      d.runStatus,
		Result:      d.result,
	}, nil
}

func (d *fakeImageGenerateDriver) Refresh(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	d.refreshHits++
	if d.mode == core.ExecutionModeAsync {
		op.Status = core.OperationStatusCompleted
		op.Result = d.result
	}
	return nil
}

func (d *fakeImageGenerateDriver) Cancel(ctx context.Context, offering catalog.Offering, op *core.Operation[imagegenerate.Result]) error {
	if d.cancelErr != nil {
		return d.cancelErr
	}
	op.Status = core.OperationStatusCanceled
	return nil
}

func TestImageGenerateExplicitProviderSyncOperation(t *testing.T) {
	driver := &fakeImageGenerateDriver{
		mode:      core.ExecutionModeSync,
		runStatus: core.OperationStatusCompleted,
		result: imagegenerate.Result{
			Images: []imagegenerate.Image{{URL: "https://example.com/image.png"}},
		},
		cancelErr: core.ErrUnsupported,
	}

	rt, err := NewBuiltins(Config{}, WithImageGenerateDriver(core.ProviderWellAPI, driver))
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	op, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelGPTImage2,
		Provider: core.ProviderWellAPI,
	}, &imagegenerate.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.Mode != core.ExecutionModeSync || op.Status != core.OperationStatusCompleted {
		t.Fatalf("unexpected operation mode/status: %+v", op)
	}
	if op.OfferingKey != "image.generate:gpt-image-2:wellapi" {
		t.Fatalf("unexpected offering key: %s", op.OfferingKey)
	}

	if err := rt.ImageGenerate().Refresh(context.Background(), op); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if driver.refreshHits != 1 {
		t.Fatalf("expected one refresh, got %d", driver.refreshHits)
	}
	if err := rt.ImageGenerate().Cancel(context.Background(), op); !errors.Is(err, core.ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

func TestImageGenerateExplicitProviderAsyncOperation(t *testing.T) {
	driver := &fakeImageGenerateDriver{
		mode:      core.ExecutionModeAsync,
		runStatus: core.OperationStatusPending,
		result: imagegenerate.Result{
			Images: []imagegenerate.Image{{URL: "https://example.com/async.png"}},
		},
	}

	rt, err := NewBuiltins(Config{}, WithImageGenerateDriver(core.ProviderKIE, driver))
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	op, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelGPTImage2,
		Provider: core.ProviderKIE,
	}, &imagegenerate.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.Mode != core.ExecutionModeAsync || op.Status != core.OperationStatusPending {
		t.Fatalf("unexpected operation mode/status: %+v", op)
	}

	if err := rt.ImageGenerate().Refresh(context.Background(), op); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if op.Status != core.OperationStatusCompleted {
		t.Fatalf("expected completed after refresh, got %s", op.Status)
	}
}

func TestImageGenerateAmbiguousProviderReturnsError(t *testing.T) {
	rt, err := NewBuiltins(Config{})
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	_, err = rt.ImageGenerate().Run(context.Background(), core.Target{
		Model: core.ModelGPTImage2,
	}, &imagegenerate.Request{Prompt: "hello"})
	if err == nil {
		t.Fatal("expected ambiguous provider error")
	}
	if !strings.Contains(err.Error(), "provider is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImageGenerateAndEditModelsLabSingleProviderRouting(t *testing.T) {
	rt, err := NewBuiltins(
		Config{},
		WithImageGenerateDriver(core.ProviderModelsLab, &fakeImageGenerateDriver{
			mode:      core.ExecutionModeAsync,
			runStatus: core.OperationStatusPending,
			result: imagegenerate.Result{
				Images: []imagegenerate.Image{{URL: "https://example.com/modelslab-flux.png"}},
			},
		}),
		WithImageEditDriver(core.ProviderModelsLab, &fakeImageEditDriver{}),
	)
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	genOp, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model: core.ModelModelsLabFlux,
	}, &imagegenerate.Request{Prompt: "flux"})
	if err != nil {
		t.Fatalf("ImageGenerate Run returned error: %v", err)
	}
	if genOp.OfferingKey != "image.generate:modelslab-flux:modelslab" {
		t.Fatalf("unexpected image generate offering: %s", genOp.OfferingKey)
	}
	if err := rt.ImageGenerate().Refresh(context.Background(), genOp); err != nil {
		t.Fatalf("ImageGenerate Refresh returned error: %v", err)
	}
	if genOp.Status != core.OperationStatusCompleted || len(genOp.Result.Images) != 1 {
		t.Fatalf("unexpected image generate operation: %+v", genOp)
	}

	interiorOp, err := rt.ImageEdit().Run(context.Background(), core.Target{
		Model: core.ModelModelsLabInterior,
	}, &imageedit.Request{Prompt: "interior", Images: []string{"https://example.com/in.png"}})
	if err != nil {
		t.Fatalf("ImageEdit interior Run returned error: %v", err)
	}
	if interiorOp.OfferingKey != "image.edit:modelslab-interior:modelslab" {
		t.Fatalf("unexpected image edit offering: %s", interiorOp.OfferingKey)
	}
	if err := rt.ImageEdit().Refresh(context.Background(), interiorOp); err != nil {
		t.Fatalf("ImageEdit interior Refresh returned error: %v", err)
	}
	if interiorOp.Status != core.OperationStatusCompleted {
		t.Fatalf("unexpected image edit interior operation: %+v", interiorOp)
	}

	exteriorOp, err := rt.ImageEdit().Run(context.Background(), core.Target{
		Model: core.ModelModelsLabExterior,
	}, &imageedit.Request{Prompt: "exterior", Images: []string{"https://example.com/in.png"}})
	if err != nil {
		t.Fatalf("ImageEdit exterior Run returned error: %v", err)
	}
	if exteriorOp.OfferingKey != "image.edit:modelslab-exterior:modelslab" {
		t.Fatalf("unexpected image edit offering: %s", exteriorOp.OfferingKey)
	}
}

func TestImageGenerateQwenFastReplicateRouting(t *testing.T) {
	rt, err := NewBuiltins(
		Config{},
		WithImageGenerateDriver(core.ProviderReplicate, &fakeImageGenerateDriver{
			mode:      core.ExecutionModeAsync,
			runStatus: core.OperationStatusPending,
			result: imagegenerate.Result{
				Images: []imagegenerate.Image{{URL: "https://example.com/qwen-fast.png"}},
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	op, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelQwenImageFast,
		Provider: core.ProviderReplicate,
	}, &imagegenerate.Request{Prompt: "fast"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.OfferingKey != "image.generate:qwen-image-fast:replicate" {
		t.Fatalf("unexpected offering key: %s", op.OfferingKey)
	}
	if err := rt.ImageGenerate().Refresh(context.Background(), op); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if op.Status != core.OperationStatusCompleted {
		t.Fatalf("expected completed after refresh, got %s", op.Status)
	}
}

func TestPortableSafetyModeMapping(t *testing.T) {
	textReq := text2ImageRequest(&imagegenerate.Request{
		Prompt:     "hello",
		SafetyMode: core.SafetyModeDisabled,
	})
	if !textReq.DisableSafetyChecker {
		t.Fatal("expected disabled safety mode to disable v1 safety checker")
	}

	videoReq := aiVideoRequest(&videogenerate.Request{
		Prompt:     "video",
		SafetyMode: core.SafetyModeDisabled,
	})
	if videoReq.NSFWChecker == nil || *videoReq.NSFWChecker {
		t.Fatal("expected disabled safety mode to map to nsfw_checker=false")
	}
}

func TestPortableGenerateAudioTriStateMapping(t *testing.T) {
	req := aiVideoRequest(&videogenerate.Request{Prompt: "video"})
	if req.GenerateAudio != nil {
		t.Fatalf("expected nil generate_audio when v2 request leaves it unset, got %#v", req.GenerateAudio)
	}

	enable := true
	req = aiVideoRequest(&videogenerate.Request{
		Prompt:        "video",
		GenerateAudio: &enable,
	})
	if req.GenerateAudio == nil || !*req.GenerateAudio {
		t.Fatalf("expected generate_audio=true, got %#v", req.GenerateAudio)
	}

	disable := false
	req = aiVideoRequest(&videogenerate.Request{
		Prompt:        "video",
		GenerateAudio: &disable,
	})
	if req.GenerateAudio == nil || *req.GenerateAudio {
		t.Fatalf("expected generate_audio=false, got %#v", req.GenerateAudio)
	}
}

func TestTextServiceForOfferingMappings(t *testing.T) {
	t.Run("replicate_qwen_image_fast", func(t *testing.T) {
		service, err := textServiceForOffering(catalog.Offering{
			Key:      "image.generate:qwen-image-fast:replicate",
			Model:    core.ModelQwenImageFast,
			Provider: core.ProviderReplicate,
		}, ProviderConfig{APIKey: "replicate-key"})
		if err != nil {
			t.Fatalf("textServiceForOffering returned error: %v", err)
		}
		if _, ok := service.(*replicate.PrunaAIQwenImageFastService); !ok {
			t.Fatalf("expected *replicate.PrunaAIQwenImageFastService, got %T", service)
		}
	})

	t.Run("modelslab_flux", func(t *testing.T) {
		service, err := textServiceForOffering(catalog.Offering{
			Key:      "image.generate:modelslab-flux:modelslab",
			Model:    core.ModelModelsLabFlux,
			Provider: core.ProviderModelsLab,
		}, ProviderConfig{APIKey: "modelslab-key"})
		if err != nil {
			t.Fatalf("textServiceForOffering returned error: %v", err)
		}
		if _, ok := service.(*modelslab.FluxService); !ok {
			t.Fatalf("expected *modelslab.FluxService, got %T", service)
		}
	})

	t.Run("kie_qwen_image", func(t *testing.T) {
		service, err := textServiceForOffering(catalog.Offering{
			Key:      "image.generate:qwen-image:kie",
			Model:    core.ModelQwenImage,
			Provider: core.ProviderKIE,
		}, ProviderConfig{APIKey: "kie-key"})
		if err != nil {
			t.Fatalf("textServiceForOffering returned error: %v", err)
		}
		if _, ok := service.(*kie.QwenText2ImageService); !ok {
			t.Fatalf("expected *kie.QwenText2ImageService, got %T", service)
		}
	})
}

func TestImageServiceForOfferingMappings(t *testing.T) {
	cases := []struct {
		name     string
		offering catalog.Offering
		wantType any
	}{
		{
			name: "replicate_nano_banana",
			offering: catalog.Offering{
				Key:      "image.edit:nano-banana:replicate",
				Model:    core.ModelNanoBanana,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.NanoBananaService{},
		},
		{
			name: "kie_nano_banana",
			offering: catalog.Offering{
				Key:      "image.edit:nano-banana:kie",
				Model:    core.ModelNanoBanana,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.NanoBananaService{},
		},
		{
			name: "replicate_controlnet",
			offering: catalog.Offering{
				Key:      "image.edit:controlnet:replicate",
				Model:    core.ModelControlNet,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.ControlNetService{},
		},
		{
			name: "modelslab_interior",
			offering: catalog.Offering{
				Key:      "image.edit:modelslab-interior:modelslab",
				Model:    core.ModelModelsLabInterior,
				Provider: core.ProviderModelsLab,
			},
			wantType: &modelslab.InteriorService{},
		},
		{
			name: "modelslab_exterior",
			offering: catalog.Offering{
				Key:      "image.edit:modelslab-exterior:modelslab",
				Model:    core.ModelModelsLabExterior,
				Provider: core.ProviderModelsLab,
			},
			wantType: &modelslab.ExteriorService{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := imageServiceForOffering(tc.offering, ProviderConfig{APIKey: "modelslab-key"})
			if err != nil {
				t.Fatalf("imageServiceForOffering returned error: %v", err)
			}
			if reflect.TypeOf(service) != reflect.TypeOf(tc.wantType) {
				t.Fatalf("expected %T, got %T", tc.wantType, service)
			}
		})
	}
}

func TestTextServiceForOfferingMappings_AllPortableImageGenerateOfferings(t *testing.T) {
	cases := []struct {
		name     string
		offering catalog.Offering
		wantType any
	}{
		{
			name: "kie_gpt_image_2",
			offering: catalog.Offering{
				Key:      "image.generate:gpt-image-2:kie",
				Model:    core.ModelGPTImage2,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.GPTImage2Text2ImageService{},
		},
		{
			name: "kie_qwen_image",
			offering: catalog.Offering{
				Key:      "image.generate:qwen-image:kie",
				Model:    core.ModelQwenImage,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.QwenText2ImageService{},
		},
		{
			name: "kie_ideogram_v3",
			offering: catalog.Offering{
				Key:      "image.generate:ideogram-v3:kie",
				Model:    core.ModelIdeogramV3,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.IdeogramV3Text2ImageService{},
		},
		{
			name: "replicate_qwen_image",
			offering: catalog.Offering{
				Key:      "image.generate:qwen-image:replicate",
				Model:    core.ModelQwenImage,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.QwenImageService{},
		},
		{
			name: "replicate_qwen_image_fast",
			offering: catalog.Offering{
				Key:      "image.generate:qwen-image-fast:replicate",
				Model:    core.ModelQwenImageFast,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.PrunaAIQwenImageFastService{},
		},
		{
			name: "replicate_flux_schnell",
			offering: catalog.Offering{
				Key:      "image.generate:flux-schnell:replicate",
				Model:    core.ModelFluxSchnell,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.FluxSchnellService{},
		},
		{
			name: "replicate_flux1dev",
			offering: catalog.Offering{
				Key:      "image.generate:flux1dev:replicate",
				Model:    core.ModelFlux1Dev,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.Flux1DevService{},
		},
		{
			name: "modelslab_flux",
			offering: catalog.Offering{
				Key:      "image.generate:modelslab-flux:modelslab",
				Model:    core.ModelModelsLabFlux,
				Provider: core.ProviderModelsLab,
			},
			wantType: &modelslab.FluxService{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := textServiceForOffering(tc.offering, ProviderConfig{APIKey: "api-key"})
			if err != nil {
				t.Fatalf("textServiceForOffering returned error: %v", err)
			}
			if reflect.TypeOf(service) != reflect.TypeOf(tc.wantType) {
				t.Fatalf("expected %T, got %T", tc.wantType, service)
			}
		})
	}
}

func TestVideoServiceForOfferingMappings(t *testing.T) {
	cases := []struct {
		name     string
		offering catalog.Offering
		wantType any
	}{
		{
			name: "kie_kling_26_image",
			offering: catalog.Offering{
				Key:      "video.generate:kling-2.6-image-to-video:kie",
				Model:    core.ModelKling26ImageToVideo,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Kling26ImageToVideoService{},
		},
		{
			name: "kie_kling_26_text",
			offering: catalog.Offering{
				Key:      "video.generate:kling-2.6-text-to-video:kie",
				Model:    core.ModelKling26TextToVideo,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Kling26TextToVideoService{},
		},
		{
			name: "kie_kling_30",
			offering: catalog.Offering{
				Key:      "video.generate:kling-3.0-video:kie",
				Model:    core.ModelKling30Video,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Kling30VideoService{},
		},
		{
			name: "kie_seedance_15_pro",
			offering: catalog.Offering{
				Key:      "video.generate:seedance-1.5-pro:kie",
				Model:    core.ModelSeedance15Pro,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Seedance15ProService{},
		},
		{
			name: "kie_seedance_2",
			offering: catalog.Offering{
				Key:      "video.generate:seedance-2:kie",
				Model:    core.ModelSeedance2,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Seedance2Service{},
		},
		{
			name: "kie_seedance_2_fast",
			offering: catalog.Offering{
				Key:      "video.generate:seedance-2-fast:kie",
				Model:    core.ModelSeedance2Fast,
				Provider: core.ProviderKIE,
			},
			wantType: &kie.Seedance2FastService{},
		},
		{
			name: "replicate_pixverse",
			offering: catalog.Offering{
				Key:      "video.generate:pixverse-v5:replicate",
				Model:    core.ModelPixverseV5,
				Provider: core.ProviderReplicate,
			},
			wantType: &replicate.PixverseV5Service{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := videoServiceForOffering(tc.offering, ProviderConfig{APIKey: "api-key"})
			if err != nil {
				t.Fatalf("videoServiceForOffering returned error: %v", err)
			}
			if reflect.TypeOf(service) != reflect.TypeOf(tc.wantType) {
				t.Fatalf("expected %T, got %T", tc.wantType, service)
			}
		})
	}
}

func TestWellAPIImageGenerateCancelUnsupported(t *testing.T) {
	driver := &wellAPIImageGenerateDriver{}
	err := driver.Cancel(context.Background(), catalog.Offering{
		Key:           "image.generate:gpt-image-2:wellapi",
		Capability:    core.CapabilityImageGenerate,
		Model:         core.ModelGPTImage2,
		Provider:      core.ProviderWellAPI,
		ExecutionMode: core.ExecutionModeSync,
	}, &core.Operation[imagegenerate.Result]{})
	if !errors.Is(err, core.ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

func readPrivateStringField(t *testing.T, v reflect.Value) string {
	t.Helper()
	ptr := unsafe.Pointer(v.UnsafeAddr())
	return reflect.NewAt(v.Type(), ptr).Elem().String()
}

func TestModelsLabAndWellAPIConfigKeysPropagate(t *testing.T) {
	rt, err := NewBuiltins(Config{
		ModelsLab: ProviderConfig{APIKey: "modelslab-key"},
		WellAPI:   ProviderConfig{APIKey: "wellapi-key"},
	})
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	modelsLabDriver, ok := rt.imageGenerateDrivers[core.ProviderModelsLab].(*modelsLabImageGenerateDriver)
	if !ok {
		t.Fatalf("expected *modelsLabImageGenerateDriver, got %T", rt.imageGenerateDrivers[core.ProviderModelsLab])
	}
	if modelsLabDriver.cfg.APIKey != "modelslab-key" {
		t.Fatalf("expected modelslab api key to propagate, got %q", modelsLabDriver.cfg.APIKey)
	}

	wellapiDriver, ok := rt.imageGenerateDrivers[core.ProviderWellAPI].(*wellAPIImageGenerateDriver)
	if !ok {
		t.Fatalf("expected *wellAPIImageGenerateDriver, got %T", rt.imageGenerateDrivers[core.ProviderWellAPI])
	}
	if wellapiDriver.cfg.APIKey != "wellapi-key" {
		t.Fatalf("expected wellapi api key to propagate, got %q", wellapiDriver.cfg.APIKey)
	}
}

func TestWellAPIImageGenerateDriverRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wellapi.PathImagesGenerations {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer wellapi-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		var payload wellapi.ImageGenerateReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if payload.Model != wellapi.ModelGPTImage2 {
			t.Fatalf("unexpected model: %s", payload.Model)
		}
		if payload.Size != "1024x1536" {
			t.Fatalf("unexpected size: %s", payload.Size)
		}

		_ = json.NewEncoder(w).Encode(wellapi.ImageGenerateResp{
			Created: 1,
			Data: []wellapi.GeneratedImage{
				{URL: "https://example.com/image.png", RevisedPrompt: "revised"},
			},
		})
	}))
	defer server.Close()

	driver := &wellAPIImageGenerateDriver{
		cfg: ProviderConfig{
			APIKey:  "wellapi-key",
			BaseURL: server.URL,
		},
	}

	op, err := driver.Run(context.Background(), catalog.Offering{
		Key:           "image.generate:gpt-image-2:wellapi",
		Capability:    core.CapabilityImageGenerate,
		Model:         core.ModelGPTImage2,
		Provider:      core.ProviderWellAPI,
		ExecutionMode: core.ExecutionModeSync,
	}, &imagegenerate.Request{
		Prompt: "hello",
		Width:  1024,
		Height: 1536,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.Status != core.OperationStatusCompleted || op.Mode != core.ExecutionModeSync {
		t.Fatalf("unexpected operation: %+v", op)
	}
	if len(op.Result.Images) != 1 || op.Result.Images[0].URL != "https://example.com/image.png" {
		t.Fatalf("unexpected result: %+v", op.Result)
	}
}

func TestRuntimeUsesCustomBaseURLForKIE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case kie.CreateTaskEndpoint:
			if got := r.Header.Get("Authorization"); got != "Bearer kie-key" {
				t.Fatalf("unexpected authorization header: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 200,
				"msg":  "success",
				"data": map[string]any{"taskId": "task-kie-1"},
			})
		case kie.RecordInfoEndpoint:
			if got := r.URL.Query().Get("taskId"); got != "task-kie-1" {
				t.Fatalf("unexpected task id query: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 200,
				"msg":  "success",
				"data": map[string]any{
					"taskId":     "task-kie-1",
					"state":      "success",
					"resultJson": `{"resultUrls":["https://example.com/kie.png"]}`,
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	rt, err := NewBuiltins(Config{
		KIE: ProviderConfig{
			APIKey:  "kie-key",
			BaseURL: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	op, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelGPTImage2,
		Provider: core.ProviderKIE,
	}, &imagegenerate.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.ExternalID != "task-kie-1" {
		t.Fatalf("unexpected external id: %s", op.ExternalID)
	}
	if err := rt.ImageGenerate().Refresh(context.Background(), op); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if op.Status != core.OperationStatusCompleted || len(op.Result.Images) != 1 {
		t.Fatalf("unexpected operation after refresh: %+v", op)
	}
}

func TestRuntimeUsesCustomBaseURLForReplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == replicate.PathPredictions:
			if got := r.Header.Get("Authorization"); got != "Bearer replicate-key" {
				t.Fatalf("unexpected authorization header: %s", got)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "pred-1",
				"status":     "starting",
				"created_at": "2026-01-01T00:00:00Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/predictions/pred-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           "pred-1",
				"status":       "succeeded",
				"output":       []string{"https://example.com/replicate.png"},
				"created_at":   "2026-01-01T00:00:00Z",
				"completed_at": "2026-01-01T00:00:02Z",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	rt, err := NewBuiltins(Config{
		Replicate: ProviderConfig{
			APIKey:  "replicate-key",
			BaseURL: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	op, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelQwenImageFast,
		Provider: core.ProviderReplicate,
	}, &imagegenerate.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if op.ExternalID != "pred-1" {
		t.Fatalf("unexpected external id: %s", op.ExternalID)
	}
	if err := rt.ImageGenerate().Refresh(context.Background(), op); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if op.Status != core.OperationStatusCompleted || len(op.Result.Images) != 1 {
		t.Fatalf("unexpected operation after refresh: %+v", op)
	}
}

func TestRuntimeUsesCustomBaseURLForModelsLab(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case modelslab.PathText2Img:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode payload: %v", err)
			}
			if payload["key"] != "modelslab-key" {
				t.Fatalf("unexpected api key payload: %#v", payload["key"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "processing",
				"id":     42,
			})
		case modelslab.PathInterior:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode payload: %v", err)
			}
			if payload["key"] != "modelslab-key" {
				t.Fatalf("unexpected api key payload: %#v", payload["key"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "processing",
				"id":     84,
			})
		case modelslab.PathFetch:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode payload: %v", err)
			}
			switch payload["request_id"] {
			case "42":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"id":     42,
					"output": []string{"https://example.com/modelslab-flux.png"},
				})
			case "84":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"id":     84,
					"output": []string{"https://example.com/modelslab-interior.png"},
				})
			default:
				t.Fatalf("unexpected request_id: %#v", payload["request_id"])
			}
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	rt, err := NewBuiltins(Config{
		ModelsLab: ProviderConfig{
			APIKey:  "modelslab-key",
			BaseURL: server.URL,
		},
	})
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	genOp, err := rt.ImageGenerate().Run(context.Background(), core.Target{
		Model: core.ModelModelsLabFlux,
	}, &imagegenerate.Request{Prompt: "flux"})
	if err != nil {
		t.Fatalf("ImageGenerate Run returned error: %v", err)
	}
	if genOp.ExternalID != "42" {
		t.Fatalf("unexpected external id: %s", genOp.ExternalID)
	}
	if err := rt.ImageGenerate().Refresh(context.Background(), genOp); err != nil {
		t.Fatalf("ImageGenerate Refresh returned error: %v", err)
	}
	if genOp.Status != core.OperationStatusCompleted || len(genOp.Result.Images) != 1 {
		t.Fatalf("unexpected generate operation after refresh: %+v", genOp)
	}

	editOp, err := rt.ImageEdit().Run(context.Background(), core.Target{
		Model: core.ModelModelsLabInterior,
	}, &imageedit.Request{
		Prompt: "interior",
		Images: []string{"https://example.com/in.png"},
	})
	if err != nil {
		t.Fatalf("ImageEdit Run returned error: %v", err)
	}
	if editOp.ExternalID != "84" {
		t.Fatalf("unexpected external id: %s", editOp.ExternalID)
	}
	if err := rt.ImageEdit().Refresh(context.Background(), editOp); err != nil {
		t.Fatalf("ImageEdit Refresh returned error: %v", err)
	}
	if editOp.Status != core.OperationStatusCompleted || len(editOp.Result.Images) != 1 {
		t.Fatalf("unexpected edit operation after refresh: %+v", editOp)
	}
}

func TestRuntimeHelperFunctions(t *testing.T) {
	if got := imageSize(&imagegenerate.Request{Width: 1024, Height: 1536}); got != "1024x1536" {
		t.Fatalf("unexpected image size: %s", got)
	}
	if got := imageSize(&imagegenerate.Request{}); got != "auto" {
		t.Fatalf("unexpected default image size: %s", got)
	}

	if got := textStatus(0); got != core.OperationStatusPending {
		t.Fatalf("unexpected text status: %s", got)
	}
	if got := imageStatus(1); got != core.OperationStatusRunning {
		t.Fatalf("unexpected image status: %s", got)
	}
	if got := videoStatus(2); got != core.OperationStatusCompleted {
		t.Fatalf("unexpected video status: %s", got)
	}

	imageResult := imageResultFromURLs([]string{"https://example.com/a.png", "", "https://example.com/b.png"})
	if len(imageResult.Images) != 2 {
		t.Fatalf("unexpected image result: %+v", imageResult)
	}
	editResult := imageEditResultFromURLs([]string{"https://example.com/edit.png"})
	if len(editResult.Images) != 1 {
		t.Fatalf("unexpected image edit result: %+v", editResult)
	}
	videoResult := videoResultFromURLs([]string{"https://example.com/video.mp4"})
	if len(videoResult.Videos) != 1 {
		t.Fatalf("unexpected video result: %+v", videoResult)
	}
}

func TestImageEditAndVideoGeneratePortableServices(t *testing.T) {
	rt, err := NewBuiltins(
		Config{},
		WithImageEditDriver(core.ProviderReplicate, &fakeImageEditDriver{}),
		WithVideoGenerateDriver(core.ProviderReplicate, &fakeVideoGenerateDriver{}),
	)
	if err != nil {
		t.Fatalf("NewBuiltins returned error: %v", err)
	}

	editOp, err := rt.ImageEdit().Run(context.Background(), core.Target{
		Model:    core.ModelControlNet,
		Provider: core.ProviderReplicate,
	}, &imageedit.Request{Prompt: "edit", Images: []string{"https://example.com/in.png"}})
	if err != nil {
		t.Fatalf("ImageEdit Run returned error: %v", err)
	}
	if editOp.OfferingKey != "image.edit:controlnet:replicate" {
		t.Fatalf("unexpected image edit offering: %s", editOp.OfferingKey)
	}
	if err := rt.ImageEdit().Refresh(context.Background(), editOp); err != nil {
		t.Fatalf("ImageEdit Refresh returned error: %v", err)
	}
	if editOp.Status != core.OperationStatusCompleted || len(editOp.Result.Images) != 1 {
		t.Fatalf("unexpected image edit operation: %+v", editOp)
	}

	videoOp, err := rt.VideoGenerate().Run(context.Background(), core.Target{
		Model:    core.ModelPixverseV5,
		Provider: core.ProviderReplicate,
	}, &videogenerate.Request{Prompt: "video"})
	if err != nil {
		t.Fatalf("VideoGenerate Run returned error: %v", err)
	}
	if videoOp.OfferingKey != "video.generate:pixverse-v5:replicate" {
		t.Fatalf("unexpected video offering: %s", videoOp.OfferingKey)
	}
	if err := rt.VideoGenerate().Refresh(context.Background(), videoOp); err != nil {
		t.Fatalf("VideoGenerate Refresh returned error: %v", err)
	}
	if videoOp.Status != core.OperationStatusCompleted || len(videoOp.Result.Videos) != 1 {
		t.Fatalf("unexpected video operation: %+v", videoOp)
	}
}
