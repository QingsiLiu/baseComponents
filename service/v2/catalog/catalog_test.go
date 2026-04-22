package catalog

import (
	"strings"
	"testing"

	v1aivideo "github.com/QingsiLiu/baseComponents/service/aivideo"
	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
	v1image2image "github.com/QingsiLiu/baseComponents/service/image2image"
	v1text2image "github.com/QingsiLiu/baseComponents/service/text2image"
	"github.com/QingsiLiu/baseComponents/service/v2/core"
)

func TestBuiltinOfferingsResolveExplicitProvider(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	offering, err := dir.Resolve(core.Target{
		Capability: core.CapabilityImageGenerate,
		Model:      core.ModelGPTImage2,
		Provider:   core.ProviderWellAPI,
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if offering.Key != "image.generate:gpt-image-2:wellapi" {
		t.Fatalf("unexpected offering key: %s", offering.Key)
	}
	if offering.ExecutionMode != core.ExecutionModeSync {
		t.Fatalf("expected sync mode, got %s", offering.ExecutionMode)
	}
}

func TestBuiltinOfferingsResolveOfferingKey(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	offering, err := dir.Resolve(core.Target{OfferingKey: "image.generate:gpt-image-2:kie"})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if offering.Provider != core.ProviderKIE {
		t.Fatalf("expected KIE provider, got %s", offering.Provider)
	}
}

func TestBuiltinOfferingsRequireProviderWhenAmbiguous(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = dir.Resolve(core.Target{
		Capability: core.CapabilityImageGenerate,
		Model:      core.ModelGPTImage2,
	})
	if err == nil {
		t.Fatal("expected ambiguous offering error")
	}
	if !strings.Contains(err.Error(), "provider is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildAndParseKey(t *testing.T) {
	key := BuildKey(core.CapabilityImageGenerate, core.ModelGPTImage2, core.ProviderKIE, "preview")
	if key != "image.generate:gpt-image-2:kie:preview" {
		t.Fatalf("unexpected key: %s", key)
	}

	capability, model, provider, variant, err := ParseKey(key)
	if err != nil {
		t.Fatalf("ParseKey returned error: %v", err)
	}
	if capability != core.CapabilityImageGenerate || model != core.ModelGPTImage2 || provider != core.ProviderKIE || variant != "preview" {
		t.Fatalf("unexpected parsed key: %s %s %s %s", capability, model, provider, variant)
	}
}

func TestBuiltinOfferingsIncludeImageAndVideoFirstPhase(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	required := []string{
		"image.generate:gpt-image-2:kie",
		"image.generate:gpt-image-2:wellapi",
		"image.generate:qwen-image:replicate",
		"image.generate:qwen-image:kie",
		"image.edit:nano-banana:replicate",
		"image.edit:nano-banana:kie",
		"video.generate:kling-2.6-image-to-video:kie",
		"video.generate:pixverse-v5:replicate",
	}

	for _, key := range required {
		if _, ok := dir.Get(key); !ok {
			t.Fatalf("expected builtin offering %s", key)
		}
	}
}

func TestBuiltinOfferingsCoverAllV1PortableImageAndVideoSources(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	expected := map[string]string{
		v1text2image.SourceReplicateFluxSchnell:          "image.generate:flux-schnell:replicate",
		v1text2image.SourceReplicateFlux1Dev:             "image.generate:flux1dev:replicate",
		v1text2image.SourceReplicateQwenImage:            "image.generate:qwen-image:replicate",
		v1text2image.SourceModelslabFlux:                 "image.generate:modelslab-flux:modelslab",
		v1text2image.SourceReplicatePrunaAIQwenImageFast: "image.generate:qwen-image-fast:replicate",
		v1text2image.SourceKieQwenImageText2Image:        "image.generate:qwen-image:kie",
		v1text2image.SourceKieIdeogramV3Text2Image:       "image.generate:ideogram-v3:kie",
		v1text2image.SourceKieGPTImage2Text2Image:        "image.generate:gpt-image-2:kie",
		v1image2image.SourceReplicateNanoBanana:          "image.edit:nano-banana:replicate",
		v1image2image.SourceReplicateControlNet:          "image.edit:controlnet:replicate",
		v1image2image.SourceModelsLabInterior:            "image.edit:modelslab-interior:modelslab",
		v1image2image.SourceModelsLabExterior:            "image.edit:modelslab-exterior:modelslab",
		v1image2image.SourceKieNanoBanana:                "image.edit:nano-banana:kie",
		v1aivideo.SourceReplicatePixverse:                "video.generate:pixverse-v5:replicate",
		v1aivideo.SourceKieKling26ImageToVideo:           "video.generate:kling-2.6-image-to-video:kie",
		v1aivideo.SourceKieKling26TextToVideo:            "video.generate:kling-2.6-text-to-video:kie",
		v1aivideo.SourceKieKling30Video:                  "video.generate:kling-3.0-video:kie",
		v1aivideo.SourceKieSeedance15Pro:                 "video.generate:seedance-1.5-pro:kie",
		v1aivideo.SourceKieSeedance2:                     "video.generate:seedance-2:kie",
		v1aivideo.SourceKieSeedance2Fast:                 "video.generate:seedance-2-fast:kie",
	}

	for source, key := range expected {
		if _, ok := dir.Get(key); !ok {
			t.Fatalf("expected v1 source %s to map to builtin offering %s", source, key)
		}
	}

	// Native-only provider-specific video capabilities stay out of the portable catalog.
	if _, ok := dir.Get("video.generate:wellapi-kling-motion-control:wellapi"); ok {
		t.Fatal("wellapi kling motion control should remain native-only")
	}
	if aivideokling.SourceWellAPIKlingMotionControl == "" || aivideokling.SourceWellAPIKlingEffects == "" {
		t.Fatal("expected aivideo/kling sources to stay defined for native usage")
	}
}

func TestBuiltinOfferingsResolveSingleProviderModelWithoutProvider(t *testing.T) {
	dir, err := New(BuiltinOfferings())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	offering, err := dir.Resolve(core.Target{
		Capability: core.CapabilityImageGenerate,
		Model:      core.ModelModelsLabFlux,
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if offering.Key != "image.generate:modelslab-flux:modelslab" {
		t.Fatalf("unexpected offering key: %s", offering.Key)
	}
}
