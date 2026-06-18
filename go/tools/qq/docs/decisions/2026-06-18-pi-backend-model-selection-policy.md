# 2026-06-18 Backend and model selection policy

**Date:** 2026-06-18

**Deciders:** Bosco Domingo

Decision: `qq` will reduce its supported agentic backend set to Claude, Cursor, OpenCode, Pi, Codex, and Gemini, then build model selection as a first-class capability across those backends.

## Context

`qq` is for quick questions. It currently supports many agentic CLIs through one static backend map. That broad backend list makes every backend-specific feature expensive to design well, test, and document.

The Pi quota issue exposed a broader problem: many supported agents can choose models, and `qq` needs a reliable way to select one without relying on sticky tool state or hard-pinning a single expensive model. Pi originally used `pi -p <prompt>`, which let Pi reuse sticky/default model state. In practice, that selected OpenAI API models and surfaced quota errors even when subscription-backed Pi usage worked in the TUI.

Hard-pinning `github-copilot/gpt-5.5` fixes that Pi quota path, but it overfits `qq` to one provider/model and makes cheap questions unnecessarily expensive.

## Decision

Limit first-class agentic backend support to:

- Claude
- Cursor
- OpenCode
- Pi
- Codex
- Gemini

Remove or demote the long-tail backend list from first-class support. Future backend additions should be deliberate and come with model-selection metadata when the backend supports model choice.

Model selection will be backend-wide, not Pi-specific. For each supported backend, `qq` should discover or query the model options exposed by the agent CLI when that is practical. When discovery is unavailable or unreliable, `qq` should use a small fallback list of popular, provider-qualified models for that backend.

Expose a user override for selecting the model per invocation. The override should pass through to the selected backend using that backend's native model argument. When provider qualification is necessary to avoid wrong billing/auth paths, the fallback/default model entries must be provider-qualified.

Use a minimal backend interface for execution, and separate capability interfaces for model discovery:

```go
// LLMInfo describes the backend-specific data for a model.
type LLMInfo struct {
	// Provider is the backend-specific name of the company, service, or lab
	// that offers the model.
	Provider string

	// ID is the backend-specific model identifier.
	ID string

	// Raw is the exact backend model selector, usually Provider + ID,
	// used when the backend expects one string or the selector cannot be
	// losslessly represented as Provider plus ID.
	Raw string
}

type Backend interface {
	Name() backendName
	Args(prompt string, model *LLMInfo) []string
	Env() []string
}

type ModelLister interface {
	ListModels(ctx context.Context) ([]LLMInfo, error)
}

type ModelProviderLister interface {
	ListModelProviders(ctx context.Context) ([]string, error)
}
```

`Provider`, `ID`, and `Raw` must preserve the backend's native spelling and casing. Do not canonicalize provider or model names in `LLMInfo`; any normalization belongs in UI/search code.

Do not add prompt-difficulty classification in this decision. Static model defaults plus explicit user override are enough for now.

## Rationale

Reducing the backend set trades breadth for depth. `qq` becomes more useful if its supported backends have reliable, documented behavior instead of a large list of shallow mappings.

Model selection belongs in the shared backend abstraction because Pi is not special here. Claude, Cursor, OpenCode, Codex, Gemini, and Pi all have model or provider concepts that users may need to control.

Discovering models from the agent when possible keeps `qq` aligned with installed tool versions. Fallback lists are still needed because not every CLI exposes a stable non-interactive model list.

Stopping short of automatic prompt routing keeps the CLI predictable and testable. This also matches `qq`'s existing design: small wrapper, explicit backend behavior, no hidden banners or complex routing.

## Options considered

1. Keep broad backend support and add Pi-only model handling.
2. Hard-pin working models for individual backends as problems appear.
3. Reduce first-class backend support and build a shared model-selection base.

## Migration Plan (if applicable)

Replace the temporary Pi hard-pin with shared model-selection support. Keep `internal/backend/backend.go` authoritative for backend mappings until a separate model registry or discovery layer exists.

Update tests to cover:

- supported backend list reduced to Claude, Cursor, OpenCode, Pi, Codex, and Gemini
- the minimal `Backend` interface renders prompts with and without model overrides
- `ModelLister` and `ModelProviderLister` are capability interfaces, not mandatory backend methods
- per-invocation model override rendering for backends that support model choice
- fallback model rendering when model discovery is unavailable
- provider-qualified model passing for Pi so it cannot silently fall back to OpenAI API billing

## Risks (if applicable)

Removing backends may break users who relied on long-tail support. This is acceptable for first-class behavior if the reduced set is clearly documented.

Model discovery can be slow, flaky, or version-dependent. Cache or keep discovery optional; never make quick prompts depend on slow startup network calls.

Fallback model lists can drift. Keep them small, obvious, and easy to update.

Automatic fallback after provider errors could hide auth/quota failures. Prefer explicit failure reporting unless a later decision accepts fallback behavior.

## Sources

- [https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/docs/providers.md](https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/docs/providers.md) — Accessed 2026-06-18; Pi supports subscription OAuth, API-key credentials, and credential resolution through auth file/env state.
- Repo state — `go/tools/qq/docs/spec.md` and `go/tools/qq/internal/backend/backend.go` define the current broad backend mapping and make backend argv rendering the authoritative source.
