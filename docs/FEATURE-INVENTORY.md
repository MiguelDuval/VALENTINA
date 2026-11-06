# VALENTINA — Phase 0 Feature Inventory

**Task ID:** VAL-P0-001
**Repository:** MiguelDuval/VALENTINA
**Branch:** `audit/valentina-phase0-feature-inventory-20261106`
**Base branch:** `chore/agent-collaboration-ci-20261106`
**Base SHA:** `a2d67f8901c949f4d1012634b573c1e08044817c`
**Huginn submodule:** `f01c1d84c08b1de972d8cc268da25602598adff8`
**Status:** Documentation only — no application code changed
**Evidence level:** SOURCE VERIFIED (all paths and symbols read from the pinned tree)

---

## 1. Executive Summary

VALENTINA is a private, local, single-operator research and CRM application for finding, checking, classifying and organizing publicly available renovation, repair, construction and interior-finishing opportunities in Valencia and owner-approved surrounding areas.

It is built by reusing the pinned `scrypster/huginn` Go AI-agent platform (submodule at `upstream/huginn`, release `v0.5.0`, commit `f01c1d84c08b1de972d8cc268da25602598adff8`) as the single application runtime, with its embedded Vue 3 / TypeScript / Vite web client rebranded as VALENTINA under ADR-005.

This document is a **source-level feature inventory**. Every entry identifies exact source paths, important symbols, actual behavior, dependencies, tests, current implementation status, proposed decision (KEEP / KEEP_BUT_MOVE / HIDE_RECOMMENDED / DISABLE_RECOMMENDED / REMOVE_LATER_CANDIDATE / OWNER_DECISION), and confidence level. No implementation changes are authorized by this task; all decisions are recommendations for a separately approved Phase 1 cleanup.

### Key findings at a glance

| Category | Count | Status |
|----------|-------|--------|
| Backend Go packages (internal/) | ~48 | SOURCE VERIFIED |
| Frontend routes (Vue router) | 13 top-level routes | SOURCE VERIFIED |
| Frontend views | 13 top-level views + subdirectories | SOURCE VERIFIED |
| AI provider backends | 10 provider IDs + 1 Managed local-runtime mode | SOURCE VERIFIED, all KEEP |
| Secret credential slots | 11 slots | SOURCE VERIFIED |
| Built-in local tools | ~16 builtins + web + git + MCP | SOURCE VERIFIED |
| MCP transport types | stdio, http | SOURCE VERIFIED |
| OAuth integration providers | 5 (Google, GitHub, Slack, Jira, Bitbucket) | SOURCE VERIFIED |
| Connection tool integrations | ~22 third-party services | SOURCE VERIFIED |
| VALENTINA domain entities | 0 implemented; 6 provisional JSON schema files | DOCUMENTED ONLY |
| Tests executed | NOT RUN (no test execution performed) | NOT RUN |

---

## 2. Repository and Runtime Architecture

### 2.1 Repository layout (VALENTINA root)

| Path | Purpose | Evidence |
|------|---------|----------|
| `AGENTS.md` | Mandatory agent instructions | SOURCE VERIFIED |
| `.clinerules/00-valentina-project.md` | Cline project operating contract | SOURCE VERIFIED |
| `docs/` | Architecture, product, security, workflow documents | SOURCE VERIFIED |
| `patches/upstream-ui/001-valentina-ui-foundation.patch` | VALENTINA UI rebranding patch (applied at build time, outside submodule) | SOURCE VERIFIED |
| `schemas/` | Provisional JSON schemas for VALENTINA domain entities | SOURCE VERIFIED |
| `scripts/` | Bootstrap, build, publish scripts | SOURCE VERIFIED |
| `templates/huginn/` | VALENTINA starter agent/skill/workflow templates (inactive) | SOURCE VERIFIED |
| `templates/owner-profile.template.yaml` | Public placeholder for owner business profile | SOURCE VERIFIED |
| `testdata/fixtures/synthetic/` | Synthetic test fixtures (example.com domain, no real data) | SOURCE VERIFIED |
| `upstream/huginn` | Pinned Git submodule — scrypster/huginn Go source | SOURCE VERIFIED |
| `.github/workflows/valentina-ui-build.yml` | CI: Windows test build, uploads VALENTINA.exe artifact | SOURCE VERIFIED |


### 2.2 Runtime architecture

One local Go application process (`scrypster/huginn`) serves an embedded Vue 3 browser frontend. The frontend is built into `internal/server/dist` and embedded into the Go executable via `-tags embed_frontend`. The application binds to `127.0.0.1` only (default port 8421).

**Startup flow (SOURCE VERIFIED):**
1. `main.go` — entry point, parses flags (`--headless`, `--cwd`, `--command`, `--dangerously-skip-permissions`, etc.)
2. `startup_wiring.go` — wires server, agent orchestrator, session store, connection manager, model store, runtime manager, notification store
3. `internal/server/server.go` — HTTP + WebSocket server, route registration, middleware
4. `internal/server/handlers.go` — API handlers
5. `internal/server/ws.go` — WebSocket message handling
6. `internal/agent/orchestrator.go` — agent execution loop
7. `internal/scheduler/scheduler.go` — cron scheduler and workflow runner
8. `internal/mcp/manager.go` — MCP server connections

**Configuration:** `~/.huginn/config.json` (atomic write, file lock). Secrets: OS keychain (Windows Credential Manager / macOS Keychain / Linux Secret Service) with `~/.huginn/secrets.json` fallback.

### 2.3 Backend package inventory (internal/)

| Package | Path | Purpose | Key symbols | Tests | Status |
|---------|------|---------|-------------|-------|--------|
| agent | internal/agent/ | Agent orchestration, chat engine, tool execution | Orchestrator, ChatEngine, AgentSession | Extensive | SOURCE VERIFIED |
| agents | internal/agents/ | Agent registry, config, toolbelt, memory | AgentsConfig, Agent, Toolbelt | Extensive | SOURCE VERIFIED |
| backend | internal/backend/ | LLM provider backends | Backend, ExternalBackend, AnthropicBackend, OpenRouterBackend, GoogleAIBackend, VertexBackend | Extensive | SOURCE VERIFIED |
| config | internal/config/ | Configuration loading/saving | Config, BackendConfig, WebUIConfig | Extensive | SOURCE VERIFIED |
| connections | internal/connections/ | OAuth connections manager | Manager, Connection | Extensive | SOURCE VERIFIED |
| mcp | internal/mcp/ | MCP client/server manager | ServerManager, MCPClient, MCPServerConfig | Extensive | SOURCE VERIFIED |
| modelconfig | internal/modelconfig/ | Model capability registry | ModelInfo, ModelRegistry, CapabilityTier | Extensive | SOURCE VERIFIED |
| models | internal/models/ | Model manifest and store | ModelEntry, Store, LockEntry | Extensive | SOURCE VERIFIED |
| permissions | internal/permissions/ | Permission gate | Gate, Decision, CheckResult | Extensive | SOURCE VERIFIED |
| scheduler | internal/scheduler/ | Cron scheduler, workflow runner | Scheduler, WorkflowRunner | Extensive | SOURCE VERIFIED |

## 3. Backend API Inventory

### 3.1 API route pattern

Routes are registered in `internal/server/server.go` (route registration via `http.HandleFunc` patterns) and `internal/server/handlers.go` (handler implementations). The API prefix is `/api/v1/`. Authentication uses a bearer token (`/api/v1/token` returns the server token; clients send `Authorization: Bearer <token>`).

### 3.2 Major API handler groups

| Handler file | Routes (prefix `/api/v1/`) | Purpose | Auth | Tests |
|--------------|---------------------------|---------|------|-------|
| `handlers.go` | token, health, restart, log level, config | Core server ops | Token | Yes |
| `handlers_agents.go` | agents/* | Agent CRUD, delegation, local tools | Token | Yes |
| `handlers_artifacts.go` | artifacts/* | Artifact CRUD | Token | Yes |
| `handlers_backend_cache.go` | backend-cache/* | Backend cache management | Token | Yes |
| `handlers_capability_matrix.go` | capability-matrix | Runtime capability reporting | Token | Yes |
| `handlers_catalog.go` | catalog/* | Credential catalog validation | Token | Yes |
| `handlers_cloud*` | cloud/* | Cloud satellite connection | Token | Yes |
| `handlers_companies.go` | companies/* | Huginn collaboration companies (NOT VALENTINA domain) | Token | Yes |
| `handlers_connections.go` | connections/* | OAuth connection management | Token | Yes |
| `handlers_config*` | config/* | Configuration read/update | Token | Yes |
| `handlers_mcp*` | mcp/* | MCP server management | Token | Yes |
| `handlers_models*` | models/* | Model catalog, pull, status | Token | Yes |
| `handlers_notification*` | notifications/* | Notification management | Token | Yes |
| `handlers_skills*` | skills/* | Skill registry, hot reload | Token | Yes |
| `handlers_spaces*` | spaces/* | Spaces, threads, replies | Token | Yes |
| `handlers_scheduler*` | scheduler/* | Scheduler, cron, workflow runs | Token | Yes |
| `handlers_workflow*` | workflows/* | Workflow CRUD, execution | Token | Yes |
| `handlers_builtin.go` | builtin/* | Built-in tool execution (bash, file, etc.) | Token | Yes |
| `handlers_session*.go` | sessions/* | Session management | Token | Yes |
| `handlers_stats*.go` | stats/* | Statistics | Token | Yes |
| `handlers_audit*` | audit/* | Audit entity log | Token | Yes |
| `handlers_secrets.go` | secrets/* | Secret status listing | Token | Yes |
| `handlers_validate_url.go` | validate-url | URL validation | Token | Yes |
| `handlers_vault_health.go` | vault-health | Vault health check | Token | Yes |
| `ws.go` | /api/v1/ws | WebSocket streaming | Token | Yes |
| `static.go` / `static_embed.go` | / | Static file serving, embedded frontend | None (public) | Yes |

### 3.3 Authentication and middleware

- Bearer token from `/api/v1/token`, sent as `Authorization: Bearer <token>`
- IP-based auth failure rate limiter (`authFailLimiter`)
- Per-endpoint rate limiters (session create, space create, workflow run, mutation)
- WebSocket origin/auth checking (`ws_auth.go`)
- `dangerously-skip-permissions` flag (CLI only, not a runtime bypass for UI)
- Permission gate: `internal/permissions/permissions.go` — per-tool allow/deny with session-level "always allow"

| secrets | internal/secrets/ | Secret storage (keychain + file) | Manager, KeyringProvider, KnownSlots | Extensive | SOURCE VERIFIED |
| server | internal/server/ | HTTP/WebSocket server, API handlers | Server, Capabilities, ValidateWiring | Extensive | SOURCE VERIFIED |
| session | internal/session/ | Session/persistent chat store | StoreInterface, SQLiteStore | Extensive | SOURCE VERIFIED |
| skills | internal/skills/ | Skill loading, registry, execution | Registry, Loader, MarkdownSkill | Extensive | SOURCE VERIFIED |
| sqlitedb | internal/sqlitedb/ | SQLite database | DB, Schema, Migrate | Extensive | SOURCE VERIFIED |
| tools | internal/tools/ | Built-in tool registry | Registry, FetchURLTool, WebSearchTool, BashTool | Extensive | SOURCE VERIFIED |
| (25 more supporting packages) | Various | artifact, storage, memory, notification, stats, runtime, workspace, tray, tui, etc. | | | SOURCE VERIFIED |

### 2.4 Frontend architecture (web/)

| File/Directory | Purpose | Key elements | Status |
|----------------|---------|--------------|--------|
| web/src/App.vue | Shell, icon-rail navigation, layout | 
avGroups, NAV_GROUPS, section switching | SOURCE VERIFIED |

## 4. Frontend Route / Navigation / View / Component Inventory

### 4.1 Vue Router routes (SOURCE VERIFIED — `web/src/router/index.ts`)

| Path | Component | Purpose | Layout section | Auth | Status |
|------|-----------|---------|----------------|------|--------|
| `/` | redirect `/chat` | Root redirect | chat | Token | SOURCE VERIFIED |
| `/chat/:sessionId?` | ChatView | Chat / assistant interface | chat | Token | SOURCE VERIFIED |
| `/space/:spaceId` | ChatView | Space chat | chat | Token | SOURCE VERIFIED |
| `/agents/:agentName?` | AgentsView | Agent management | admin | Token | SOURCE VERIFIED |
| `/models/:provider?` | ModelsView | Model configuration | admin | Token | SOURCE VERIFIED |
| `/connections` | ConnectionsView | OAuth connections | admin | Token | SOURCE VERIFIED |
| `/settings` | SettingsView | Application settings | admin | Token | SOURCE VERIFIED |
| `/logs` | LogsView | Application logs | logs | Token | SOURCE VERIFIED |
| `/stats` | StatsView | Statistics | logs | Token | SOURCE VERIFIED |
| `/inbox` | InboxView | Activity log / notifications | logs | Token | SOURCE VERIFIED |
| `/memory` | MemoryView | Memory vault | admin | Token | SOURCE VERIFIED |
| `/routines` | redirect `/workflows` | Alias to workflows | automation | Token | SOURCE VERIFIED |
| `/workflows/:id?` | WorkflowsView | Workflow management | automation | Token | SOURCE VERIFIED |
| `/workflows/:id/runs/:runId` | WorkflowsView | Workflow run detail | automation | Token | SOURCE VERIFIED |
| `/cloud` | CloudView | Cloud satellite status | (own section) | Token | SOURCE VERIFIED |
| `/skills/:tab?` | SkillsView | Skill management | admin | Token | SOURCE VERIFIED |
| `/:pathMatch(.*)*` | redirect `/` | Catch-all redirect | chat | Token | SOURCE VERIFIED |

**Total: 13 unique routes + catch-all**

### 4.2 Navigation groups (SOURCE VERIFIED — `web/src/utils/navLayout.ts`)

| Group | Items | Paths |
|-------|-------|-------|
| chat | Chat | `/chat` |
| admin | Agents, Memory, Models, Automation (Workflows), Connections, Skills, Settings | `/agents`, `/memory`, `/models`, `/workflows`, `/connections`, `/skills/browse`, `/settings` |
| logs | Stats, Logs, Activity Log (Inbox) | `/stats`, `/logs`, `/inbox` |

**Default route:** `/chat` (NOT Overview/Review Queue — this is a VALENTINA customization gap; the patch adds an Overview view but the upstream router still redirects root to `/chat`).

### 4.3 Frontend views (SOURCE VERIFIED — `web/src/views/`)

| View | File | Subdirectory | Purpose | Data dependency |
|------|------|--------------|---------|-----------------|
| AgentsView | `AgentsView.vue` | `agents/` | Agent management, roster | `/api/v1/agents/*` |
| ChatView | `ChatView.vue` | `chat/` | Chat interface, streaming | WebSocket `/api/v1/ws` |
| CloudView | `CloudView.vue` | | Cloud satellite status | `/api/v1/cloud/*` |
| ConnectionsView | `ConnectionsView.vue` | `connections/` | OAuth connections | `/api/v1/connections/*` |
| InboxView | `InboxView.vue` | | Activity log, notifications | `/api/v1/notifications/*` |
| LogsView | `LogsView.vue` | | Application logs | `/api/v1/logs/*` |
| MemoryView | `MemoryView.vue` | | Memory vault | `/api/v1/memory/*` |
| ModelsView | `ModelsView.vue` | `models/` | Model configuration | `/api/v1/models/*` |
| SessionsView | `SessionsView.vue` | `chat/` | Session management | `/api/v1/sessions/*` |
| SettingsView | `SettingsView.vue` | `settings/` | Application settings | `/api/v1/config/*` |
| SkillsView | `SkillsView.vue` | | Skill management | `/api/v1/skills/*` |
| StatsView | `StatsView.vue` | | Statistics | `/api/v1/stats/*` |
| WorkflowsView | `WorkflowsView.vue` | `workflows/` | Workflow management | `/api/v1/workflows/*` |

**Total: 13 top-level views**

### 4.4 Key frontend composables (SOURCE VERIFIED — `web/src/composables/`)

| Composable | File | Purpose | API dependency |
|------------|------|---------|----------------|
| useApi | `useApi.ts` | HTTP API client with bearer token auth, WebSocket, SSE streaming | All `/api/v1/*` |
| useAgents | `useAgents.ts` | Agent CRUD, delegation | `/api/v1/agents/*` |
| useWorkflows | `useWorkflows.ts` | Workflow CRUD, runs | `/api/v1/workflows/*` |
| useSkills | `useSkills.ts` | Skill registry | `/api/v1/skills/*` |
| useConnections | `useConnections.ts` | OAuth connections | `/api/v1/connections/*` |
| useConfig | `useConfig.ts` | Configuration read/update | `/api/v1/config/*` |
| useFirstRun | `useFirstRun.ts` | First-run setup | `/api/v1/*` |
| useCloud | `useCloud.ts` | Cloud satellite | `/api/v1/cloud/*` |
| useHuginnWS | `useHuginnWS.ts` | WebSocket connection | `/api/v1/ws` |
| useSessions | `useSessions.ts` | Session management | `/api/v1/sessions/*` |
| useModels | `useModels.ts` | Model management | `/api/v1/models/*` |

### 4.5 Frontend test evidence

- `web/src/views/__tests__/` — 15 test files (AgentsView, App, ChatView, CloudView, ConnectionsView, InboxView, LogsView, MemoryView, ModelsView, SessionsView, SettingsView, SkillsView, StatsView, WorkflowsView)
- `web/src/composables/__tests__/` — composable tests
- `web/src/components/__tests__/` — component tests
- Test runner: Vitest (inferred from `vitest` imports in patch test)
- **Status:** SOURCE VERIFIED, TESTS EXIST, NOT EXECUTED in this task

## 5. Full AI Provider / Model / Compute Inventory

### 5.1 Provider registration (SOURCE VERIFIED — `internal/backend/factory.go`)

The `NewFromConfig` / `newFromResolvedConfig` function switches on provider name:

| # | Provider | Config value | Implementation file | Constructor | Default endpoint | Decision |
|---|----------|-------------|---------------------|-------------|------------------|----------|
| 1 | Ollama (local) | `"ollama"` or `""` | `internal/backend/external.go` | `NewExternalBackend(endpoint)` | `http://localhost:11434` | **KEEP** |
| 2 | External (generic OpenAI-compatible) | `"external"` | `internal/backend/external.go` | `NewExternalBackend(endpoint)` | `http://localhost:11434` | **KEEP** |
| 3 | Custom (generic OpenAI-compatible, requires endpoint) | `"custom"` | `internal/backend/external.go` | `NewExternalBackendWithAPIKey(endpoint, resolver)` | Required (no default) | **KEEP** — used for NVIDIA NIM via OpenAI-compatible endpoint |
| 4 | OpenAI | `"openai"` | `internal/backend/external.go` | `NewExternalBackendWithAPIKey(endpoint, resolver)` | `https://api.openai.com` | **KEEP** |
| 5 | DeepSeek | `"deepseek"` | `internal/backend/external.go` | `NewExternalBackendWithAPIKey(endpoint, resolver)` | `https://api.deepseek.com` | **KEEP** |
| 6 | Z.AI | `"zai"` | `internal/backend/external.go` | `NewExternalBackendWithAPIKey(endpoint, resolver)` | `https://api.z.ai/api/paas/v4` | **KEEP** |
| 7 | Anthropic | `"anthropic"` | `internal/backend/anthropic.go` | `NewAnthropicBackend(resolver, model)` | `https://api.anthropic.com` | **KEEP** |
| 8 | OpenRouter | `"openrouter"` | `internal/backend/openrouter.go` | `NewOpenRouterBackend(resolver, modelID)` | `https://openrouter.ai/api/v1` | **KEEP** |
| 9 | Google AI (AI Studio) | `"google"` | `internal/backend/google_ai.go` | `NewGoogleAIBackend(resolver, model)` | `https://generativelanguage.googleapis.com` | **KEEP** |
| 10 | Vertex AI | `"vertex"` | `internal/backend/vertex.go` | `NewVertexBackend(ctx, VertexConfig{...})` | GCP regional endpoint | **KEEP** |

### 5.2 Backend interface (SOURCE VERIFIED — `internal/backend/backend.go`)

```go
type Backend interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Health(ctx context.Context) error
    Shutdown(ctx context.Context) error
    ContextWindow() int
}
```

`StatusReporter` interface for circuit-breaker state reporting (Anthropic, Google AI, Vertex implement it).


| web/src/main.ts | Vue application entry | App mounting, plugin registration | SOURCE VERIFIED |
| web/src/router/index.ts | Vue router configuration | 13 routes, hash history | SOURCE VERIFIED |
| web/src/utils/navLayout.ts | Route>layout mapping, nav groups | AppSection, NAV_GROUPS, TOP_LEVEL_SECTIONS | SOURCE VERIFIED |
| web/src/style.css | Tailwind directives, base styles | VALENTINA color tokens (via patch) | SOURCE VERIFIED |
| web/src/composables/ | Vue composables (API client, state) | useApi, useAgents, useWorkflows, useSkills, useConnections, useConfig, useFirstRun, useCloud, useHuginnWS, etc. | SOURCE VERIFIED |
| web/src/components/ | Shared Vue components | Agent cards, chat threads, modals, pickers | SOURCE VERIFIED |
| web/src/views/ | Page-level views | AgentsView, ChatView, ModelsView, ConnectionsView, SettingsView, LogsView, StatsView, InboxView, WorkflowsView, SkillsView, MemoryView, CloudView | SOURCE VERIFIED |
| web/src/views/__tests__/ | Frontend unit tests | 15 test files | SOURCE VERIFIED |



### 5.3 Managed backend (local inference)

| Provider | Config value | Implementation | Notes | Decision |
|----------|-------------|----------------|-------|----------|
| Managed (built-in llama.cpp) | `"managed"` | `internal/backend/managed.go` (selected and initialized by `main.go`; separate runtime mode, NOT a case in `internal/backend/factory.go`) | Phase 3 — local model subprocess management; `main.go` selects and initializes managed mode | **KEEP** (implemented, configurable) |

### 5.4 Key resolution (SOURCE VERIFIED — `internal/backend/keyring.go`, `internal/secrets/manager.go`)

| Mechanism | Description | Code path | Decision |
|-----------|-------------|-----------|----------|
| `$ENV_VAR` | Read environment variable | `ResolveAPIKey()` in `keyring.go`; `Resolve()` in `secrets/manager.go` | **KEEP** |
| `keyring:service:slot` | OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service) | `keyring.Get()` via `go-keyring`; `Manager.Resolve()` | **KEEP** |
| Literal value | Raw API key string (not recommended, flagged by `IsLiteralAPIKey`) | Returned as-is | **KEEP** (but flagged) |
| Empty | No key configured | Returns empty string | — |

### 5.5 Known secret slots (SOURCE VERIFIED — `internal/secrets/manager.go`)

`KnownSlots` is declared in `internal/secrets/manager.go` (not `provider.go`, which only defines the `KeyringProvider` interface):

`KnownSlots` = `["anthropic", "openai", "openrouter", "deepseek", "zai", "brave", "google", "github", "slack", "jira", "bitbucket"]`

**11 credential slots**, each independently configurable. Storage priority: OS keychain → `~/.huginn/secrets.json` plaintext JSON fallback with 0600 file permissions.


### 5.6 Model discovery and catalogs

| Catalog | File | Contents | Source | Decision |
|---------|------|----------|--------|----------|
| Curated local models | `internal/models/models.json` (embedded) | qwen2.5-coder:7b, qwen2.5-coder:3b, qwen2.5-coder:14b (GGUF, local inference) | Embedded at compile time | **KEEP** |
| Provider model catalog | `internal/models/provider_catalog.json` (embedded) | Anthropic: 8 models (4 deprecated); OpenAI: 8 models (3 deprecated) | Embedded at compile time, CDN overlay supported | **KEEP** |
| User model manifest | `~/.huginn/models.user.json` (optional) | User-added local models | Local file, optional | **KEEP** |
| Model lock store | `~/.huginn/models.lock.json` | Installed local model records | Local file | **KEEP** |

### 5.7 Model capability metadata

| Capability | Implementation | Evidence | Decision |
|------------|---------------|----------|----------|
| Context window | `modelconfig.ModelInfo.ContextWindow`, `backend.ContextWindowForVertexModel()` | `internal/modelconfig/registry.go`, `internal/backend/vertex.go` | **KEEP** |
| Supports tools | `modelconfig.ModelInfo.SupportsTools`, `UnreliableForTools()` | `internal/modelconfig/registry.go` | **KEEP** |
| Supports vision | `backend.DetectVision()`, `visionFromShow()` (Ollama model family detection) | `internal/backend/capabilities.go` | **KEEP** (Ollama only; heuristic) |
| Capability tier | `modelconfig.CapabilityTier` (high/medium/low) | `internal/modelconfig/registry.go` | **KEEP** |
| Streaming | All backends implement SSE streaming via `ChatCompletion` | `internal/backend/*` | **KEEP** |
| Tool/function calling | All backends support `ChatRequest.Tools` | `internal/backend/backend.go` | **KEEP** |
| Structured output | Via tool calling + JSON parsing | Application layer | **KEEP** |
| Embeddings | NOT FOUND in source | — | DOCUMENTED ONLY |
| Model discovery | `provider_catalog.go` `GlobalProviderCatalog()`, `models.go` `LoadMerged()` | `internal/models/` | **KEEP** |
| Custom model IDs | All backends accept arbitrary model ID strings | `factory.go` | **KEEP** |

### 5.8 Per-agent provider/model selection

| Feature | Implementation | File | Decision |
|---------|---------------|------|----------|
| Per-agent model override | `agents.Agent` config has `ModelID` field | `internal/agents/config.go` | **KEEP** |
| Per-agent provider binding | Via backend config + agent model reference | `internal/agents/` | **KEEP** |
| Model registry lookup | `ModelRegistry.Lookup()`, `ModelSupportsTools()` | `internal/modelconfig/registry.go` | **KEEP** |

### 5.9 Provider comparison table (exhaustive)

| Provider/backend | Implementation path | Endpoint configuration | Credential mechanism | Model selection | Agent-level support | Capabilities | Tests | UI location | Decision |
|------------------|--------------------|----------------------|---------------------|-----------------|-------------------|--------------|-------|-------------|----------|
| Ollama (local LLM) | `internal/backend/external.go` | `http://localhost:11434` (default, configurable) | None (local) | Model ID string | Yes (per-agent model ID) | Streaming, tool calling (live probe), vision (family detection) | `external_test.go`, `capabilities_test.go` | Settings → Models | **KEEP** |
| External (OpenAI-compatible) | `internal/backend/external.go` | Configurable (default `http://localhost:11434`) | Optional | Model ID string | Yes | Streaming, tool calling | `external_test.go` | Settings → Models | **KEEP** |
| Custom (NIM-compatible) | `internal/backend/external.go` | Required (user-configured) | `$ENV_VAR` or `keyring:svc:user` | Model ID string | Yes | Streaming, tool calling | `external_test.go` | Settings → Models | **KEEP** |
| OpenAI | `internal/backend/external.go` | `https://api.openai.com` | API key | Model ID string | Yes | Streaming, tools, vision | `external_test.go` | Settings → Models | **KEEP** |
| DeepSeek | `internal/backend/external.go` | `https://api.deepseek.com` | API key | Model ID string | Yes | Streaming, tools | `external_test.go` | Settings → Models | **KEEP** |
| Z.AI | `internal/backend/external.go` | `https://api.z.ai/api/paas/v4` | API key | Model ID string | Yes | Streaming | `external_test.go` | Settings → Models | **KEEP** |
| Anthropic | `internal/backend/anthropic.go` | `https://api.anthropic.com` | API key | Model ID string | Yes | Streaming, tools, vision, extended thinking | `anthropic_test.go` | Settings → Models | **KEEP** |
| OpenRouter | `internal/backend/openrouter.go` | `https://openrouter.ai/api/v1` | API key | Model ID string | Yes | Streaming, tools, provider routing | `openrouter_test.go` | Settings → Models | **KEEP** |
| Google AI Studio | `internal/backend/google_ai.go` | `https://generativelanguage.googleapis.com` | API key (query param) | gemini-* models | Yes | Streaming, tools | `google_ai_test.go` | Settings → Models | **KEEP** |
| Vertex AI | `internal/backend/vertex.go` | GCP regional | GCP OAuth / service account | gemini-* / claude-* | Yes | Streaming, tools, vision | `vertex_test.go` | Settings → Models | **KEEP** |
| Managed (llama.cpp) | `internal/backend/managed.go` | Local subprocess | None | Model filename | Yes | Local inference | Selected by `main.go` (separate runtime mode; not a case in `internal/backend/factory.go`) | Settings → Models | **KEEP** |

**All 10 verified provider backends are classified KEEP.** No provider is removed, restricted, hidden, or made NIM-exclusive.


## 6. MCP Inventory

### 6.1 MCP client (SOURCE VERIFIED — `internal/mcp/`)

| Component | File | Purpose | Key symbols | Tests | Decision |
|-----------|------|---------|-------------|-------|----------|
| Client | `internal/mcp/client.go` | MCP JSON-RPC client | `MCPClient`, `Initialize`, `ListTools`, `CallTool` | `client_test.go` | **KEEP** |
| Manager | `internal/mcp/manager.go` | MCP server lifecycle, circuit breaker, reconnection | `ServerManager`, `StartAll`, `StopAll`, `CallToolGated` | `manager_test.go` | **KEEP** |
| Config | `internal/mcp/config.go` | MCP server configuration | `MCPServerConfig`, `ApprovalGate` | `config_test.go` | **KEEP** |
| Types | `internal/mcp/types.go` | MCP protocol types | `MCPTool`, `MCPToolCallResult`, `ProtocolVersion` | `types_test.go` | **KEEP** |
| Stdio transport | `internal/mcp/stdio.go` | stdio-based MCP server transport | `NewStdioTransport` | `stdio_test.go` | **KEEP** |
| HTTP transport | `internal/mcp/http_transport.go` | HTTP-based MCP server transport | `NewHTTPTransport` | `http_transport_test.go` | **KEEP** |
| Transport | `internal/mcp/transport.go` | Transport interface | `Transport` interface | | **KEEP** |

### 6.2 Transport types

| Transport | File | Direction | Credential handling | Approval gate | Decision |
|-----------|------|-----------|-------------------|---------------|----------|
| stdio | `internal/mcp/stdio.go` | Client spawns local process | Via `cfg.Env` array (may contain secrets) | Default true (fail-closed) | **KEEP** |
| http | `internal/mcp/http_transport.go` | Client connects to HTTP endpoint | Via `BEARER_TOKEN=` env var | Default true (fail-closed) | **KEEP** |

### 6.3 MCP tool gating

- `CallToolGated()` — circuit breaker with failure threshold (3 failures → open for 10s)
- `ApprovalGate` — per-server default `true` (fail-closed; operator must explicitly opt out with `"approval_gate": false`)
- Tool name collision guards on registration and reconnect
- Per-agent toolbelt integration

### 6.4 No VALENTINA-specific MCP adapter exists

No source-policy-enforcing MCP adapter is implemented. The `SOURCE-REGISTRY.md` document identifies this gap: a VALENTINA fetch/search MCP adapter that enforces allowed_hosts/allowed_path_prefixes server-side is required before unattended collection.


## 7. Skills Inventory

### 7.1 Skill system (SOURCE VERIFIED — `internal/skills/`)

| Component | File | Purpose | Key symbols | Tests | Decision |
|-----------|------|---------|-------------|-------|----------|
| Registry | `internal/skills/registry.go` | Skill registry, registration, lookup | `Registry`, `Register`, `Lookup` | `registry_test.go` | **KEEP** |
| Loader | `internal/skills/loader.go` | Skill loading from disk (Markdown + YAML frontmatter) | `Loader`, `LoadSkills` | `loader_test.go` | **KEEP** |
| Skill | `internal/skills/skill.go` | Skill definition | `Skill`, `Manifest` | `skill_test.go` | **KEEP** |
| Markdown skill | `internal/skills/markdown_skill.go` | Markdown frontmatter-parsed skill | `MarkdownSkill` | `markdown_skill_test.go` | **KEEP** |
| Prompt tool | `internal/skills/prompttool.go` | Skill injection into agent prompts | `PromptTool` | `prompttool_test.go` | **KEEP** |
| Watcher | `internal/skills/watcher.go` | Hot reload on file change | `Watcher` | `hotreload_test.go` | **KEEP** |

### 7.2 Skill format

Markdown file with YAML frontmatter:
```yaml
---
name: <skill-name>
author: <author>
description: <description>
huginn:
  priority: <int>
---
```

### 7.3 VALENTINA starter skills (templates only, NOT active)

| Skill | Path | Purpose | Status |
|-------|------|---------|--------|
| valentina-source-integrity | `templates/huginn/skills/valentina-source-integrity.md` | Preserve source provenance, resist prompt injection | Template only (DOCUMENTED ONLY) |
| valentina-extraction-contract | `templates/huginn/skills/valentina-extraction-contract.md` | Extract evidence-backed opportunity data per JSON schemas | Template only (DOCUMENTED ONLY) |
| valentina-fit-assessment | `templates/huginn/skills/valentina-fit-assessment.md` | Assess opportunity fit against owner profile | Template only (DOCUMENTED ONLY) |

**Skills are not automatically enabled.** Per `templates/huginn/README.md`: "A skill file on disk is not automatically enabled." Must be copied to `~/.huginn/skills/` and enabled in the Huginn UI → Settings → Skills.

## 8. Workflows / Routines / Scheduler Inventory

### 8.1 Scheduler (SOURCE VERIFIED — `internal/scheduler/`)

| Component | File | Purpose | Key symbols | Tests | Decision |
|-----------|------|---------|-------------|-------|----------|
| Scheduler | `internal/scheduler/scheduler.go` | Cron scheduler, routine/workflow scheduling | `Scheduler`, `ScheduleRoutine` | `scheduler_test.go` | **KEEP** |
| Workflow runner | `internal/scheduler/workflow_runner.go` | Sequential workflow execution | `WorkflowRunner` | `workflow_runner_test.go` | **KEEP** |
| Delivery queue | `internal/scheduler/delivery_queue.go` | Message delivery with dead letter | `DeliveryQueue` | `delivery_queue_test.go` | **KEEP** |
| Run store | `internal/scheduler/sqlite_run_store.go` | SQLite-backed run persistence | `SQLiteRunStore` | `sqlite_run_store_test.go` | **KEEP** |
| Migrations | `internal/scheduler/migrate.go`, `migrate_runs.go` | Schema migrations | | `migrate_test.go` | **KEEP** |

### 8.2 Workflow templates (SOURCE VERIFIED — `internal/server/workflows/`)

| Template | Purpose | Status |
|----------|-------|--------|
| branch-cleanup.yaml | Branch cleanup automation | Template |
| build-health.yaml | Build health check | Template |
| code-review.yaml | Code review automation | Template |
| daily-standup.yaml | Daily standup | Template |
| dependency-updates.yaml | Dependency updates | Template |
| health-check.yaml | Health check | Template |
| morning-standup.yaml | Morning standup | Template |
| stale-pr-check.yaml | Stale PR check | Template |
| weekly-summary.yaml | Weekly summary | Template |

### 8.3 Routine templates (SOURCE VERIFIED — `internal/server/routines/`)

5 routine YAML templates (same names as workflows).

### 8.4 Automation templates (SOURCE VERIFIED — `internal/server/automations/`)

5 automation YAML templates.

### 8.5 VALENTINA starter workflow (template only, NOT active)

| Workflow | Path | Trigger | Status |
|----------|------|---------|--------|
| valentina-research-manual | `templates/huginn/workflows/valentina-research-manual.yaml` | Manual (disabled) | Template only (DOCUMENTED ONLY) — 3 steps: discover (Scout), verify (Verifier), assess (Assessor). `enabled: false`, `mode: manual`. Does NOT write to a business database. |

#### Scout/Verifier tool enforcement distinction (SOURCE VERIFIED)

The VALENTINA-Scout agent template (`templates/huginn/agents/valentina-scout.json`) declares `local_tools: ["web_search", "fetch_url"]`. Its system prompt instructs the agent to operate within "explicit owner-approved scope and source policy." However:

- **Prompt-level instruction:** The Scout's system prompt requests approved-source behavior — this is guidance to the LLM, not a technical authorization boundary.
- **Template-defined tool configuration:** The template configures `web_search` and `fetch_url` as available local tools.
- **Actually provisioned/active agent:** The agent template is not automatically installed or activated — it must be copied to `~/.huginn/agents/` and enabled in the Huginn UI.
- **Server-side technical enforcement:** The pinned Huginn `fetch_url` (`internal/tools/fetch_url.go`) provides SSRF protection (blocks private/internal IPs), a 10-second timeout, and a 512 KB response limit. It does **not** receive a VALENTINA source-profile ID and does **not** enforce the source registry's `allowed_hosts` or `allowed_path_prefixes` fields.

**Conclusion:** The source registry (`docs/SOURCE-REGISTRY.md`, `schemas/source-profile.schema.json`) is a policy document only. It is **not** a technical authorization boundary in the pinned Huginn release. Unattended collection must remain blocked until a VALENTINA source-policy MCP adapter is implemented and tested.

### 8.6 Scheduler behavior on restart

Per `QUALITY-AND-OPERATIONS.md`: "Huginn's scheduler runs inside the Huginn process. Scheduled routines/workflows run only while it is running; missed schedules are not automatically replayed."


## 9. Chat / Sessions / Inbox / Artifacts / Logs Inventory

| Feature | Path | Purpose | Storage | Tests | Decision |
|---------|------|---------|---------|-------|----------|
| Chat | `internal/agent/chat_engine.go`, `web/src/views/ChatView.vue` | Real-time chat with streaming | Session store (SQLite) | Extensive | **KEEP** |
| Sessions | `internal/session/sqlite_store.go` | Persistent chat sessions | SQLite | Extensive | **KEEP** |
| Inbox | `internal/server/handlers_notification.go`, `web/src/views/InboxView.vue` | Notifications, activity log | Notification store (Pebble/SQLite) | Extensive | **KEEP** |
| Artifacts | `internal/artifact/` | Reviewable outputs (not permanent business DB) | | Extensive | **KEEP** |
| Logs | `internal/logger/`, `web/src/views/LogsView.vue` | Application logging | | | **KEEP** |
| Memory vault | `internal/agents/memory.go`, `web/src/views/MemoryView.vue` | Agent long-term memory | SQLite | Extensive | **KEEP** |
| Stats | `internal/stats/`, `web/src/views/StatsView.vue` | Runtime statistics | | | **KEEP** |
| Spaces | `internal/spaces/` | Multi-threaded collaboration spaces | | | **KEEP** |
| Notifications | `internal/notification/` | Notification store with pruning | Pebble + SQLite | | **KEEP** |

## 10. Data / Configuration / Storage Inventory

### 10.1 Storage locations

| Location | Purpose | Git-ignored? | Contains secrets? | Decision |
|----------|---------|-------------|-------------------|----------|
| `~/.huginn/config.json` | Application configuration | Yes (local) | API key references (not values) | **KEEP** |
| `~/.huginn/secrets.json` | Secret fallback store | Yes | Plaintext JSON map of slot→value, written with 0600 file permissions (not application-encrypted) | **KEEP** |
| OS keychain | Primary secret storage | N/A | Actual API keys, tokens | **KEEP** |
| `~/.huginn/models/` | Local model files (GGUF) | Yes | No | **KEEP** |
| `~/.huginn/models.lock.json` | Model install lock | Yes | No | **KEEP** |
| `~/.huginn/agents/` | Agent configurations | Yes | No | **KEEP** |
| `~/.huginn/skills/` | Skill files | Yes | No | **KEEP** |
| `~/.huginn/workflows/` | Workflow definitions | Yes | No | **KEEP** |
| `~/.huginn/store/notifications/` | Notification store (Pebble) | Yes | No | **KEEP** |

### 10.2 VALENTINA domain entities (NOT implemented — schemas only)

| Entity | Schema file | Status | Notes |
|--------|-------------|--------|-------|
| Company | `schemas/company.schema.json` | Provisional JSON contract only | No SQLite table, no API, no UI |
| Opportunity | `schemas/opportunity.schema.json` | Provisional JSON contract only | demand_evidence_type enum included |
| SourceEvidence | `schemas/source-evidence.schema.json` | Provisional JSON contract only | retrieval_status enum |
| SourceProfile | `schemas/source-profile.schema.json` | Provisional JSON contract only | approval_status: pending_review/approved/blocked/paused |
| OwnerProfile | `schemas/owner-profile.schema.json` | Provisional JSON contract only | Local-only, never committed |
| WorkflowInput | `schemas/workflow-input.schema.json` | Provisional JSON contract only | Manual research trigger |

**Domain entity implementation status: 0 implemented.** Six provisional JSON schema files exist, but no VALENTINA domain entity is implemented in database, API, or UI persistence:

- **Company, Opportunity, SourceEvidence** — have corresponding record-oriented schemas (with required fields, enums, and validation rules).
- **SourceProfile** — a source-policy/profile contract, not automatically proof of a complete Source domain implementation.
- **OwnerProfile** — a supporting business-profile contract (local-only, never committed).
- **WorkflowInput** — a supporting manual-research trigger contract.
- **ResearchRun** and **ReviewDecision** — dedicated schema contracts are **missing** unless new source evidence proves otherwise.

No SQLite tables, API handlers, or UI routes exist for any VALENTINA domain entity.

### 10.3 Synthetic test fixtures

| File | Path | Purpose |
|------|------|---------|
| company.synthetic.json | `testdata/fixtures/synthetic/company.synthetic.json` | Example.com domain, fictional |
| opportunity.synthetic.json | `testdata/fixtures/synthetic/opportunity.synthetic.json` | Example.com domain, fictional |
| source-evidence.synthetic.json | `testdata/fixtures/synthetic/source-evidence.synthetic.json` | Example.com domain, fictional |

## 11. Security and Permissions

### 11.1 Permission gate (SOURCE VERIFIED — `internal/permissions/permissions.go`)

| Component | Purpose | Key behaviors |
|-----------|---------|---------------|
| `Gate` | Controls tool call execution | `skipAll` (`--dangerously-skip-permissions`), `watchedProviders`, per-agent toolbelt, session allow-list (LRU, max 1000 entries) |
| `Decision` | Allow / AllowOnce / AllowAll / Deny | `AllowAll` = always allow for session |
| `CheckResult` | Machine-readable decision | `Allowed`, `ReasonCode`, `Reason` |
| `PermissionRequest` | Tool call permission request | ToolName, Level, Args, Summary, Provider, AgentName, SessionID |

### 11.2 Tool permission levels (SOURCE VERIFIED — `internal/tools/types.go`)

Read tools (fetch_url, read_file, list_dir, grep, search_files, web_search) vs write tools (bash, write_file, edit_file, git_commit, git_push, etc.).

### 11.3 Security boundaries

| Boundary | Implementation | Status |
|----------|---------------|--------|
| Loopback binding | `config.WebUIConfig.Bind` defaults to `"127.0.0.1"` | SOURCE VERIFIED |
| SSRF protection | `internal/tools/fetch_url.go` `isBlockedIP()`, `ssrfSafeDialContext()` | SOURCE VERIFIED |
| fetch_url limits | 10s timeout, 512KB max response, HTML-to-Markdown only | SOURCE VERIFIED |
| No JS execution | fetch_url does not execute JavaScript | SOURCE VERIFIED |
| Circuit breaker | MCP server failure threshold | SOURCE VERIFIED |
| Approval gate | MCP `ApprovalGate` defaults to true | SOURCE VERIFIED |
| Permission prompts | TUI/WS permission request flow | SOURCE VERIFIED |
| Credential redaction | `ResolveAPIKey` avoids including resolved secret values in its own error messages | NOT VERIFIED for application-wide log redaction |
| Audit log | `internal/server/audit.go`, `audit_entity_log.go` | SOURCE VERIFIED |

### 11.4 Source registry enforcement gap

**CRITICAL FINDING:** The source registry (`docs/SOURCE-REGISTRY.md`, `schemas/source-profile.schema.json`) is a **policy document only**. The pinned Huginn `fetch_url` does NOT enforce `allowed_hosts` or `allowed_path_prefixes`. No VALENTINA MCP adapter exists to enforce this server-side. Unattended collection must not be enabled until this adapter is implemented.


## 12. Build / Test / CI Inventory

### 12.1 Build pipeline

| Step | Tool | File | Status |
|------|------|------|--------|
| Submodule init | git | `scripts/bootstrap.ps1`, `scripts/bootstrap.sh` | SOURCE VERIFIED |
| Pin verification | git | `scripts/build-valentina.ps1` | SOURCE VERIFIED |
| Patch check/apply | git | `scripts/build-valentina.ps1` | SOURCE VERIFIED |
| Frontend install | npm | `npm ci --no-audit --no-fund` | SOURCE VERIFIED |
| Frontend tests | npm | `npm test` | SOURCE VERIFIED, NOT EXECUTED |
| Frontend build | npm | `npm run build` | SOURCE VERIFIED, NOT EXECUTED |
| Go build | Go | `go build -tags embed_frontend -o VALENTINA.exe` | SOURCE VERIFIED, NOT EXECUTED |
| Go version | Go 1.25+ | From `upstream/huginn/go.mod` | SOURCE VERIFIED |
| Node version | Node.js 22 | CI uses node-version "22" | SOURCE VERIFIED |

### 12.2 CI workflow

| Workflow | File | Triggers | Artifact | Retention |
|----------|------|----------|----------|-----------|
| VALENTINA Windows test build | `.github/workflows/valentina-ui-build.yml` | push/PR on main/feature/fix/chore/docs/audit | `VALENTINA-windows-x64-${{ github.sha }}` | 30 days |

**CI workflow security description (SOURCE VERIFIED — `.github/workflows/valentina-ui-build.yml`):**

- `permissions: contents: read` — least-privilege, no token write access.
- Checkout uses `persist-credentials: false` — no credential persistence.
- Does **not** reference `NVIDIA_NIM_API_KEY` or any `secrets.*` — no AI provider API credential is used by CI.
- Does **not** push branches or merge PRs — CI is read-only with respect to Git.
- Actions references use version tags (`actions/checkout@v4`, `actions/setup-node@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`) — **not** full commit-SHA pins.
- `npm ci` and Go module/dependency setup can access package registries — this is dependency-fetch network access, distinct from web research, scraping or external business actions.
- No artifact is uploaded on failed or cancelled runs.

**CI status:** Verified via GitHub Actions:

- Push run #35: https://github.com/MiguelDuval/VALENTINA/actions/runs/42214548633 — event: push, head SHA: c00aea224f91db09cfeaaf6ad1b18584b63ef46c, branch: audit/valentina-phase0-feature-inventory-20261106, status: **Success**, artifact: `VALENTINA-windows-x64-c00aea224f91db09cfeaaf6ad1b18584b63ef46c`.
- Pull request run #36: https://github.com/MiguelDuval/VALENTINA/actions/runs/42214558815 — event: pull_request (PR #4), head SHA: c00aea224f91db09cfeaaf6ad1b18584b63ef46c, branch: audit/valentina-phase0-feature-inventory-20261106, status: **Success**, artifact: `VALENTINA-windows-x64-d0e85a0bc3eb320dd56b63efc67124d9ea436194`.

Older runs #33/#34 remain as historical evidence only (previous HEAD 78977db).

### 12.3 Test suites

| Suite | Location | Status |
|-------|----------|--------|
| Go backend tests | `internal/*/*_test.go` | SOURCE VERIFIED, NOT EXECUTED |
| Frontend unit tests | `web/src/*/__tests__/` | SOURCE VERIFIED, NOT EXECUTED |

### 12.4 Not yet executed

- `go test ./...` — NOT RUN
- `npm test` (frontend) — NOT RUN
- `go build -tags embed_frontend` — NOT RUN
- Live NIM API test — NOT RUN (no credentials)
- Windows install/launch test — NOT RUN

## 13. Feature Classification Summary

| Decision | Count | Examples |
|----------|-------|----------|
| **KEEP** | ~100+ | All AI providers, MCP, skills, workflows, scheduler, diagnostics, local APIs, security controls, chat, sessions, inbox, logs, stats, models, connections |
| **KEEP_BUT_MOVE** | Recommended | Cloud view → Advanced area; developer tools (bash, git) → Advanced area; raw fetch_url → behind source adapter |
| **HIDE_RECOMMENDED** | Recommended | CloudView from ordinary navigation; company/channel/DM rail from non-chat views; developer console surfaces |
| **DISABLE_RECOMMENDED** | Recommended | `--dangerously-skip-permissions` flag in production builds; unattended raw fetch_url without source adapter |
| **REMOVE_LATER_CANDIDATE** | None recommended at Phase 0 | No code removal recommended without owner approval |
| **OWNER_DECISION** | Recommended | Cloud/satellite connection; OAuth integrations (GitHub, Google, Slack, Jira, Bitbucket) |

### Verified providers — ALL KEEP

**10 provider IDs** (all classified **KEEP**): Ollama (`"ollama"`/`""`), External (`"external"`), Custom (`"custom"` — the OpenAI-compatible backend used for NVIDIA NIM), OpenAI (`"openai"`), DeepSeek (`"deepseek"`), Z.AI (`"zai"`), Anthropic (`"anthropic"`), OpenRouter (`"openrouter"`), Google AI (`"google"`), Vertex AI (`"vertex"`).

**1 separate Managed local-runtime mode** (also **KEEP**): `BackendConfig.Type = "managed"`, implemented via `internal/backend/managed.go`, manages a local `llama-server` subprocess lifecycle. This is a distinct backend runtime mode, not an additional case in the provider factory switch in `internal/backend/factory.go`.

No provider ID or Managed mode is removed, restricted, hidden, or made NIM-exclusive.

## 14. Missing Evidence / Unresolved Issues

| Item | Evidence level | Notes |
|------|---------------|-------|
| Live NIM compatibility | NOT VERIFIED | No credentials, no live request executed |
| Frontend test execution | NOT RUN | Test files exist but not executed |
| Go test execution | NOT RUN | Test files exist but not executed |
| Windows build | NOT RUN | Build script exists but not executed |
| Windows install/launch | NOT RUN | Documented in DEPLOYMENT-AND-INSTALLER.md |
| Source adapter enforcement | NOT IMPLEMENTED | Gap identified in SOURCE-REGISTRY.md |
| SQLite domain persistence | NOT IMPLEMENTED | Schemas exist, no physical database |
| VALENTINA CRM UI routes | NOT IMPLEMENTED | No Opportunities/Companies/Sources/Research views |
| CI artifact URL | Available | Push run #35 artifact: `VALENTINA-windows-x64-c00aea224f91db09cfeaaf6ad1b18584b63ef46c` |

## 15. Confidence Levels

All path and symbol references in this document are **SOURCE VERIFIED** (read directly from the pinned repository tree). Behavioral claims about unexecuted code paths are marked **NOT VERIFIED** or **DOCUMENTED ONLY**. No claim is made about runtime behavior, live API integration, or Windows installation without actual test evidence.


## Appendix A: Exact Source Paths Reference

**Backend providers:**
- `upstream/huginn/internal/backend/factory.go` — provider switch
- `upstream/huginn/internal/backend/external.go` — ExternalBackend (Ollama, external, custom, openai, deepseek, zai)
- `upstream/huginn/internal/backend/anthropic.go` — AnthropicBackend
- `upstream/huginn/internal/backend/openrouter.go` — OpenRouterBackend
- `upstream/huginn/internal/backend/google_ai.go` — GoogleAIBackend
- `upstream/huginn/internal/backend/vertex.go` — VertexBackend
- `upstream/huginn/internal/backend/backend.go` — Backend interface, StreamEvent types
- `upstream/huginn/internal/backend/keyring.go` — ResolveAPIKey, StoreAPIKey
- `upstream/huginn/internal/backend/capabilities.go` — ModelCapabilities, DetectVision, FetchCapabilities

**Secrets:**
- `upstream/huginn/internal/secrets/provider.go` — KeyringProvider, OSKeyring
- `upstream/huginn/internal/secrets/manager.go` — Secret Manager (keychain + file), KnownSlots (11 slots)
- `upstream/huginn/internal/secrets/file_store.go` — FileStore fallback

**Configuration:**
- `upstream/huginn/internal/config/config.go` — Config, BackendConfig, WebUIConfig, IntegrationsConfig, CloudConfig

**Models:**
- `upstream/huginn/internal/models/models.json` — curated local models (3 qwen2.5-coder variants)
- `upstream/huginn/internal/models/provider_catalog.json` — provider model catalog (Anthropic + OpenAI)
- `upstream/huginn/internal/models/provider_catalog.go` — ProviderCatalog singleton, CDN overlay
- `upstream/huginn/internal/models/manifest.go` — LoadMerged, user manifest
- `upstream/huginn/internal/models/store.go` — Model store, lock file

**MCP:**
- `upstream/huginn/internal/mcp/manager.go` — ServerManager, circuit breaker
- `upstream/huginn/internal/mcp/config.go` — MCPServerConfig, ApprovalGate
- `upstream/huginn/internal/mcp/client.go` — MCPClient
- `upstream/huginn/internal/mcp/types.go` — Protocol types
- `upstream/huginn/internal/mcp/stdio.go` — Stdio transport
- `upstream/huginn/internal/mcp/http_transport.go` — HTTP transport

**Tools:**
- `upstream/huginn/internal/tools/builtin.go` — RegisterBuiltins, RegisterWebTools, BuiltinToolNames
- `upstream/huginn/internal/tools/fetch_url.go` — FetchURLTool, SSRF protection
- `upstream/huginn/internal/tools/web_search.go` — WebSearchTool (Brave)
- `upstream/huginn/internal/tools/registry.go` — Tool registry

**Frontend:**
- `upstream/huginn/web/src/router/index.ts` — 13 routes
- `upstream/huginn/web/src/utils/navLayout.ts` — NAV_GROUPS, TOP_LEVEL_SECTIONS
- `upstream/huginn/web/src/App.vue` — Shell, navigation
- `upstream/huginn/web/src/composables/useApi.ts` — API client

**Permissions:**
- `upstream/huginn/internal/permissions/permissions.go` — Gate, Decision, CheckResult

**Server:**
- `upstream/huginn/internal/server/server.go` — Server struct, route registration
- `upstream/huginn/internal/server/handlers.go` — API handlers
- `upstream/huginn/internal/server/capabilities.go` — Server capabilities map
- `upstream/huginn/internal/server/static_embed.go` — Embedded frontend

**Startup:**
- `upstream/huginn/main.go` — Entry point, buildProviders (OAuth)
- `upstream/huginn/startup_wiring.go` — Server wiring

**VALENTINA-specific:**
- `patches/upstream-ui/001-valentina-ui-foundation.patch` — UI rebranding patch
- `templates/huginn/agents/valentina-{scout,verifier,assessor}.json` — Agent templates
- `templates/huginn/skills/valentina-{source-integrity,extraction-contract,fit-assessment}.md` — Skill templates
- `templates/huginn/workflows/valentina-research-manual.yaml` — Workflow template (disabled)
- `schemas/*.schema.json` — 6 provisional JSON schemas
- `templates/owner-profile.template.yaml` — Owner profile template
- `scripts/build-valentina.ps1` — Build script (pin verification, patch, frontend build, Go build)
- `.github/workflows/valentina-ui-build.yml` — CI workflow
