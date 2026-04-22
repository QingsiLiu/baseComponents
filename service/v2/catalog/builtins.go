package catalog

import "github.com/QingsiLiu/baseComponents/service/v2/core"

func BuiltinOfferings() []Offering {
	return []Offering{
		builtin(core.CapabilityImageGenerate, core.ModelGPTImage2, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelGPTImage2, core.ProviderWellAPI, "", core.ExecutionModeSync),
		builtin(core.CapabilityImageGenerate, core.ModelQwenImage, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelQwenImage, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelQwenImageFast, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelFluxSchnell, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelFlux1Dev, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelModelsLabFlux, core.ProviderModelsLab, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageGenerate, core.ModelIdeogramV3, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageEdit, core.ModelNanoBanana, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageEdit, core.ModelNanoBanana, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageEdit, core.ModelControlNet, core.ProviderReplicate, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageEdit, core.ModelModelsLabInterior, core.ProviderModelsLab, "", core.ExecutionModeAsync),
		builtin(core.CapabilityImageEdit, core.ModelModelsLabExterior, core.ProviderModelsLab, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelKling26ImageToVideo, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelKling26TextToVideo, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelKling30Video, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelSeedance15Pro, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelSeedance2, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelSeedance2Fast, core.ProviderKIE, "", core.ExecutionModeAsync),
		builtin(core.CapabilityVideoGenerate, core.ModelPixverseV5, core.ProviderReplicate, "", core.ExecutionModeAsync),
	}
}

func builtin(capability core.Capability, model core.Model, provider core.Provider, variant string, mode core.ExecutionMode) Offering {
	return Offering{
		Key:           BuildKey(capability, model, provider, variant),
		Capability:    capability,
		Model:         model,
		Provider:      provider,
		Variant:       variant,
		ExecutionMode: mode,
		Stable:        true,
	}
}
