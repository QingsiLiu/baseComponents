package core

// Capability identifies a portable AI capability.
type Capability string

const (
	CapabilityImageGenerate Capability = "image.generate"
	CapabilityImageEdit     Capability = "image.edit"
	CapabilityVideoGenerate Capability = "video.generate"
)

// Provider identifies the service provider that serves a model.
type Provider string

const (
	ProviderKIE       Provider = "kie"
	ProviderWellAPI   Provider = "wellapi"
	ProviderReplicate Provider = "replicate"
	ProviderModelsLab Provider = "modelslab"
)

// Model is the canonical logical model name used by v2 routing.
type Model string

const (
	ModelGPTImage2           Model = "gpt-image-2"
	ModelQwenImage           Model = "qwen-image"
	ModelQwenImageFast       Model = "qwen-image-fast"
	ModelFluxSchnell         Model = "flux-schnell"
	ModelFlux1Dev            Model = "flux1dev"
	ModelModelsLabFlux       Model = "modelslab-flux"
	ModelModelsLabInterior   Model = "modelslab-interior"
	ModelModelsLabExterior   Model = "modelslab-exterior"
	ModelIdeogramV3          Model = "ideogram-v3"
	ModelNanoBanana          Model = "nano-banana"
	ModelControlNet          Model = "controlnet"
	ModelKling26ImageToVideo Model = "kling-2.6-image-to-video"
	ModelKling26TextToVideo  Model = "kling-2.6-text-to-video"
	ModelKling30Video        Model = "kling-3.0-video"
	ModelSeedance15Pro       Model = "seedance-1.5-pro"
	ModelSeedance2           Model = "seedance-2"
	ModelSeedance2Fast       Model = "seedance-2-fast"
	ModelPixverseV5          Model = "pixverse-v5"
)

// Target describes which offering should serve a request.
type Target struct {
	Capability  Capability `json:"capability"`
	Model       Model      `json:"model"`
	Provider    Provider   `json:"provider"`
	OfferingKey string     `json:"offering_key"`
}

// ExecutionMode distinguishes synchronous and asynchronous provider behavior.
type ExecutionMode string

const (
	ExecutionModeSync  ExecutionMode = "sync"
	ExecutionModeAsync ExecutionMode = "async"
)

// OperationStatus is the provider-independent operation status.
type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "pending"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusCompleted OperationStatus = "completed"
	OperationStatusCanceled  OperationStatus = "canceled"
	OperationStatusFailed    OperationStatus = "failed"
)

// Operation is the unified result envelope for sync and async executions.
type Operation[T any] struct {
	OfferingKey string          `json:"offering_key"`
	ExternalID  string          `json:"external_id"`
	Mode        ExecutionMode   `json:"mode"`
	Status      OperationStatus `json:"status"`
	Result      T               `json:"result"`
	Raw         []byte          `json:"raw,omitempty"`
}

// SafetyMode expresses safety filtering intent without provider-specific names.
type SafetyMode string

const (
	SafetyModeDefault  SafetyMode = ""
	SafetyModeStrict   SafetyMode = "strict"
	SafetyModeDisabled SafetyMode = "disabled"
)

func (s SafetyMode) IsDisabled() bool {
	return s == SafetyModeDisabled
}
