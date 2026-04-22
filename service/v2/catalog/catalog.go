package catalog

import (
	"fmt"
	"strings"

	"github.com/QingsiLiu/baseComponents/service/v2/core"
)

// Offering describes a concrete model served by a concrete provider.
type Offering struct {
	Key           string             `json:"key"`
	Capability    core.Capability    `json:"capability"`
	Model         core.Model         `json:"model"`
	Provider      core.Provider      `json:"provider"`
	Variant       string             `json:"variant,omitempty"`
	ExecutionMode core.ExecutionMode `json:"execution_mode"`
	Stable        bool               `json:"stable"`
	NativeOnly    bool               `json:"native_only"`
}

// Directory keeps all offerings and resolves user targets.
type Directory struct {
	offerings []Offering
	byKey     map[string]Offering
}

func New(offerings []Offering) (*Directory, error) {
	d := &Directory{
		offerings: make([]Offering, 0, len(offerings)),
		byKey:     make(map[string]Offering, len(offerings)),
	}
	for _, offering := range offerings {
		if offering.Key == "" {
			offering.Key = BuildKey(offering.Capability, offering.Model, offering.Provider, offering.Variant)
		}
		if _, ok := d.byKey[offering.Key]; ok {
			return nil, fmt.Errorf("duplicate offering key: %s", offering.Key)
		}
		d.offerings = append(d.offerings, offering)
		d.byKey[offering.Key] = offering
	}
	return d, nil
}

func BuildKey(capability core.Capability, model core.Model, provider core.Provider, variant string) string {
	parts := []string{string(capability), string(model), string(provider)}
	if strings.TrimSpace(variant) != "" {
		parts = append(parts, strings.TrimSpace(variant))
	}
	return strings.Join(parts, ":")
}

func ParseKey(key string) (core.Capability, core.Model, core.Provider, string, error) {
	parts := strings.Split(key, ":")
	if len(parts) != 3 && len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid offering key %q", key)
	}
	var variant string
	if len(parts) == 4 {
		variant = parts[3]
	}
	return core.Capability(parts[0]), core.Model(parts[1]), core.Provider(parts[2]), variant, nil
}

func (d *Directory) List() []Offering {
	out := make([]Offering, len(d.offerings))
	copy(out, d.offerings)
	return out
}

func (d *Directory) Get(key string) (Offering, bool) {
	offering, ok := d.byKey[key]
	return offering, ok
}

func (d *Directory) Resolve(target core.Target) (Offering, error) {
	if strings.TrimSpace(target.OfferingKey) != "" {
		offering, ok := d.byKey[target.OfferingKey]
		if !ok {
			return Offering{}, fmt.Errorf("offering %q not found", target.OfferingKey)
		}
		return offering, nil
	}
	if target.Capability == "" {
		return Offering{}, fmt.Errorf("capability is required")
	}
	if target.Model == "" {
		return Offering{}, fmt.Errorf("model is required")
	}

	matches := make([]Offering, 0, 2)
	for _, offering := range d.offerings {
		if offering.Capability != target.Capability || offering.Model != target.Model {
			continue
		}
		if target.Provider != "" && offering.Provider != target.Provider {
			continue
		}
		matches = append(matches, offering)
	}
	if len(matches) == 0 {
		return Offering{}, fmt.Errorf("offering not found for capability=%s model=%s provider=%s", target.Capability, target.Model, target.Provider)
	}
	if len(matches) > 1 {
		return Offering{}, fmt.Errorf("ambiguous offering for capability=%s model=%s: provider is required", target.Capability, target.Model)
	}
	return matches[0], nil
}
