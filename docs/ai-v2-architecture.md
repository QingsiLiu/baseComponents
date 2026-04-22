# AI v2 Architecture

## Overview

`service/v2` is the portable image/video layer that runs alongside v1.

Goals:

- keep v1 public APIs stable
- support multiple providers for the same logical model
- unify sync and async execution through `Operation[T]`
- reuse existing provider implementations from `service/thirdparty/*`

Current scope:

- `image.generate`
- `image.edit`
- `video.generate`

LLM is intentionally still out of scope.

## Design Principles

- route by `capability + model + provider`, not by v1 `source`
- treat the same model on different providers as different offerings
- require explicit provider selection for multi-provider models
- keep portable APIs provider-neutral
- move provider-exclusive capabilities into `service/v2/native/...`

## Layering

### `service/v2/core`

Defines the shared concepts:

- `Capability`
- `Provider`
- `Model`
- `Target`
- `ExecutionMode`
- `OperationStatus`
- `Operation[T]`
- `SafetyMode`

`Target.OfferingKey` is the exact-match escape hatch.

### `service/v2/catalog`

Maintains the built-in offering directory.

Each `Offering` includes:

- capability
- canonical model
- provider
- variant
- execution mode
- stability metadata

Routing rules:

1. `OfferingKey` resolves directly.
2. Otherwise resolve by `Capability + Model + Provider`.
3. If a model exists on multiple providers and provider is omitted, resolution fails explicitly.

### `service/v2/runtime`

Responsible for:

- building the built-in catalog
- wiring provider configs
- selecting the provider wrapper
- translating portable requests into v1 provider requests
- normalizing sync and async execution into `Operation[T]`

Runtime does not use `init()` registration.

### `service/v2/native`

Holds provider-specific APIs that should not pollute portable contracts.

Current example:

- `service/v2/native/wellapi/kling`

## Operation Model

Portable services return `Operation[T]`.

Fields:

- `OfferingKey`
- `ExternalID`
- `Mode`
- `Status`
- `Result`
- `Raw`

Behavior:

- sync provider:
  - `Run` may return `completed` immediately
  - `Refresh` is effectively idempotent
  - `Cancel` returns unsupported when meaningless
- async provider:
  - `Run` returns `pending` or `running`
  - `Refresh` polls external task state
  - `Cancel` delegates to provider cancellation if supported

## Portable Coverage

### `image.generate`

- `gpt-image-2`
  - `kie`
  - `wellapi`
- `qwen-image`
  - `replicate`
  - `kie`
- `qwen-image-fast`
  - `replicate`
- `flux-schnell`
  - `replicate`
- `flux1dev`
  - `replicate`
- `modelslab-flux`
  - `modelslab`
- `ideogram-v3`
  - `kie`

### `image.edit`

- `nano-banana`
  - `replicate`
  - `kie`
- `controlnet`
  - `replicate`
- `modelslab-interior`
  - `modelslab`
- `modelslab-exterior`
  - `modelslab`

### `video.generate`

- `kling-2.6-image-to-video`
  - `kie`
- `kling-2.6-text-to-video`
  - `kie`
- `kling-3.0-video`
  - `kie`
- `seedance-1.5-pro`
  - `kie`
- `seedance-2`
  - `kie`
- `seedance-2-fast`
  - `kie`
- `pixverse-v5`
  - `replicate`

### Native-only

- WellAPI Kling motion-control
- WellAPI Kling effects

These remain outside portable `video.generate`.

## Request Semantics

### Safety

Portable APIs use `SafetyMode` instead of provider-specific names like:

- `disable_safety_checker`
- `nsfw_checker`

Mapping happens inside runtime/provider wrappers.

### Audio

`video.generate.Request.GenerateAudio` is a `*bool`.

Meaning:

- `nil`: preserve provider default behavior
- `true`: explicitly enable
- `false`: explicitly disable

This is required because some providers, such as KIE Seedance, default audio generation to `true` when the field is unset.

## Provider Config and BaseURL

`runtime.Config` is grouped by provider:

- `KIE`
- `WellAPI`
- `Replicate`
- `ModelsLab`

Each group supports:

- `APIKey`
- `BaseURL`

Rule:

- if v2 exposes `BaseURL`, runtime must pass it into the real provider client

This matters for:

- staging
- `httptest`
- alternate gateways
- mock environments

The current implementation now propagates `BaseURL` for KIE, WellAPI, Replicate, and ModelsLab.

## Reuse Strategy

v2 does not rewrite provider logic.

Instead it reuses v1 provider services:

- KIE
- Replicate
- ModelsLab
- WellAPI

That means v2 is mainly responsible for:

- routing
- request translation
- operation normalization
- config propagation

## Testing Strategy

Testing is split into three levels:

### Catalog tests

- offering key parsing
- ambiguous route failures
- v1 coverage completeness checks

### Runtime tests

- fake-driver tests for generic operation semantics
- real wrapper mapping tests for provider selection
- `BaseURL` propagation tests using `httptest`
- request semantic tests such as audio tri-state behavior

### Provider tests

- existing v1 provider tests remain the regression baseline

## Extension Workflow

When adding a new image/video model:

1. Add canonical `Model` constant in `core`.
2. Add built-in `Offering` in `catalog`.
3. Decide portable vs native.
4. Reuse or add the provider service constructor.
5. Ensure `BaseURL` can propagate through the provider client if HTTP-based.
6. Add:
   - catalog coverage test
   - runtime routing test
   - real wrapper mapping test
   - `BaseURL` test if applicable
7. Update README and the conventions doc.
