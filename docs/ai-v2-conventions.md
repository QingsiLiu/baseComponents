# AI v2 Development Conventions

## Purpose

This document defines how to extend and maintain `service/v2`.

Use it whenever you:

- add a model
- add a provider
- add a native-only capability
- change portable request semantics
- change runtime routing or provider config behavior

## Naming

### Canonical model names

Use logical model names, not v1 source identifiers.

Good examples:

- `gpt-image-2`
- `qwen-image`
- `qwen-image-fast`
- `modelslab-flux`
- `nano-banana`

Avoid embedding provider names in canonical model names unless the model is inherently provider-exclusive.

### Offering keys

Format:

`<capability>:<model>:<provider>[:<variant>]`

Examples:

- `image.generate:gpt-image-2:kie`
- `image.generate:gpt-image-2:wellapi`
- `image.edit:modelslab-interior:modelslab`

## Routing Rules

- `OfferingKey` always wins.
- Otherwise route by `Capability + Model + Provider`.
- If multiple providers serve the same model and provider is omitted, fail explicitly.
- Never add implicit provider preference in portable APIs.

## Portable vs Native

### Put a feature in portable API only if:

- its request fields make sense across multiple providers
- its result can be normalized without awkward loss of meaning
- it is part of the shared product contract

### Put a feature in native API if:

- it is provider-exclusive
- it needs provider-only fields with no clean cross-provider equivalent
- forcing it into portable APIs would make the contract confusing

Current example:

- WellAPI Kling motion-control/effects stay in `service/v2/native/wellapi/kling`

## Request Field Rules

### Preserve tri-state semantics when provider defaults matter

Use `*bool` instead of `bool` when providers distinguish:

- unset
- true
- false

Current example:

- `video.generate.Request.GenerateAudio`

Do not collapse those three states into one boolean if provider defaults depend on omission.

### Keep portable APIs minimal

Do not add provider-specific tuning flags to portable requests just to avoid writing a native package.

If only one provider meaningfully uses a field, it probably belongs in `service/v2/native/...`.

### Safety

Portable APIs use `SafetyMode`.

Do not leak provider field names like:

- `disable_safety_checker`
- `nsfw_checker`

Map them inside runtime/provider wrappers.

## Provider Config Rules

If `runtime.Config` exposes `BaseURL`, runtime must pass it into the real provider client.

Do not leave `BaseURL` as a dead config field.

Why:

- staging support
- mockability
- `httptest`
- alternate gateways

If a provider client cannot accept custom `BaseURL`, fix the client first or remove the field from the public v2 contract.

## Runtime Rules

### Reuse v1 provider services where practical

Preferred strategy:

- keep HTTP/protocol logic in `service/thirdparty/*`
- let v2 runtime handle routing, translation, and normalization

### Keep runtime decomposition tidy

As `runtime.go` grows, prefer extracting:

- provider client builders
- offering-to-service selectors
- request/result mapping helpers

Avoid piling more behavior into long switch blocks without helper boundaries.

## Testing Rules

Every new portable model/provider should have:

1. catalog coverage test
2. route resolution test
3. real wrapper mapping test
4. `BaseURL` propagation test when HTTP-based
5. sync/async operation behavior test

Native packages should have at least:

- constructor tests
- config propagation tests

Do not rely only on fake drivers when runtime correctness depends on real wrapper selection.

## Documentation Rules

Any meaningful v2 change should update at least one of:

- `README.md`
- `docs/ai-v2-architecture.md`
- `docs/ai-v2-conventions.md`

Update README when:

- provider coverage changes
- user-facing examples change
- native-only boundaries change

Update architecture doc when:

- routing changes
- operation semantics change
- a new layer or package is introduced

Update conventions doc when:

- a bug teaches a new rule
- a migration pattern becomes a reusable standard

## Checklist For Adding a Model

- Add canonical `Model` constant.
- Add `Offering` entry.
- Decide portable vs native.
- Reuse or add provider service constructor.
- Ensure `BaseURL` support is real.
- Add catalog test.
- Add runtime routing test.
- Add real wrapper mapping test.
- Add README entry.
- Update docs if the pattern is new.
