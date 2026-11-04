# Optional integrations and tools

**Purpose:** Keep the optional stack small. This document records the approved options and the conditions for introducing them. A tool listed here is not automatically installed, enabled, or authorized to act.

## 1. NVIDIA NIM — initial model API

**Role:** Hosted inference for agent reasoning, classification, extraction assistance and (with a supported model) vision.

Huginn's audited source contains a generic OpenAI-compatible `custom` backend. NVIDIA NIM exposes an OpenAI-compatible `POST /v1/chat/completions` endpoint, including streaming and tool-calling support at the API level.

Proposed Huginn configuration shape:

```json
{
  "backend": {
    "type": "external",
    "provider": "custom",
    "endpoint": "https://integrate.api.nvidia.com/v1",
    "api_key": "$NVIDIA_NIM_API_KEY"
  },
  "default_model": "<MODEL_ID_AVAILABLE_TO_THIS_ACCOUNT>"
}
```

Rules:
- Select the model ID from the account's available NIM models; do not guess or hard-code one into the application.
- Use a model that supports the capabilities required by the task. Tool calling and vision capability vary by model; the endpoint alone does not guarantee every model supports them.
- Keep the key in a protected local environment variable or OS keychain. The GitHub Actions secret name remains `NVIDIA_NIM_API_KEY`; never store a real key in Git, a prompt, a screenshot or an issue.
- Do not confuse the hosted API with downloadable/self-hosted NIM containers; they have different requirements and entitlements.
- Huginn/NIM compatibility is source-supported but still needs a real request before it can be reported as operational.

Official references:
- [NVIDIA NIM LLM API overview](https://docs.api.nvidia.com/nim/reference/llm-apis)
- [NIM API reference — chat completions](https://docs.nvidia.com/nim/large-language-models/latest/api-reference.html)
- [Hosted API base URL](https://integrate.api.nvidia.com)

## 2. Playwright MCP — preferred optional browser tool

**Role:** Browser automation for dynamic JavaScript websites when Huginn's ordinary HTTP `fetch_url` is insufficient.

The official Playwright MCP server exposes browser actions to MCP clients using structured accessibility snapshots. It can navigate pages, inspect content, click, type, fill forms and capture screenshots. It is more directly relevant to web-page automation than UFO² and should be considered first for a permitted dynamic website.

Official documentation currently requires Node.js 20 or newer and an MCP-capable client. The upstream package is `@playwright/mcp`.

Example Huginn MCP configuration shape (local configuration only; do not commit it blindly):

```json
{
  "mcp_servers": [
    {
      "name": "playwright",
      "command": "npx",
      "args": ["@playwright/mcp@<PINNED_VERSION>"],
      "transport": "stdio",
      "approval_gate": true
    }
  ]
}
```

Rules:
- Pin an explicit package version when adopting it; do not use `@latest` for a reproducible working installation.
- Before enabling it in Huginn on Windows, confirm that the configured stdio command resolves correctly on Windows (for example, `npx.cmd` may need an explicit executable/wrapper path).
- Give a discovery agent only the browser tools it needs. Do not permit form submission, purchases, messaging, job applications or other outward-facing actions without explicit human approval.
- Do not load private browser profiles, personal sessions or saved authentication state by default.
- Treat page text and accessibility snapshots as untrusted source data. Enforce permitted-source scope, rate limits and human review.
- Playwright is a browser-control layer, not a canonical lead database, deduplication engine or replacement for source-specific structured extraction.

Official references:
- [Playwright MCP installation](https://playwright.dev/mcp/installation)
- [Playwright MCP introduction](https://playwright.dev/mcp/introduction)
- [Playwright MCP source repository](https://github.com/microsoft/playwright-mcp)

**Adoption rule:** do not install it merely because it is available. Add it when the first approved target source demonstrably needs browser rendering or interaction.

## 3. Microsoft UFO² — optional Windows desktop executor

**Role:** Desktop automation for a task that genuinely requires interacting with Windows applications or native desktop UI. This is not the default way to scrape websites.

Official UFO² is located under the `ufo/` directory of the [Microsoft UFO repository](https://github.com/microsoft/UFO/tree/main/ufo). Its current quick-start documentation lists Windows 10 or newer and Python 3.10 or newer, with Git for cloning. UFO² uses a HostAgent/AppAgent architecture and supports hybrid Windows UI Automation/API actions.

Official references:
- [UFO² quick start](https://microsoft.github.io/UFO/getting_started/quick_start_ufo2/)
- [UFO² overview](https://microsoft.github.io/UFO/ufo2/overview/)
- [UFO² source README](https://github.com/microsoft/UFO/blob/main/ufo/README.md)
- [UFO² and MCP integration notes](https://microsoft.github.io/UFO/ufo2/as_galaxy_device/)

Rules:
- Keep UFO² in a separate Python virtual environment and its own installation directory. Do not copy it into the Huginn submodule or VALENTINA repository.
- Do not install or configure it until a specific Windows-only workflow justifies it.
- UFO² and Huginn are separate applications. This audit has not established a turnkey, stable Huginn-to-UFO² bridge. If required, expose only a narrow local MCP/CLI bridge and require explicit approval for side-effecting actions.
- Do not assume an NVIDIA NIM model/provider is compatible with UFO² without checking UFO²'s current model configuration.
- Do not use UFO² as a substitute for API, HTTP or Playwright-based web collection.

**Adoption rule:** optional and deferred. Use only for a real desktop-only task, not as part of every website search.

## 4. Brave Search — optional search provider

Huginn's built-in `web_search` calls the Brave Search API and requires a separate Brave API key. It is independent of NVIDIA NIM; the NIM key does not authorize Brave Search.

Keep the Brave key in protected local configuration. Enable this only when this search source and its account/usage terms are acceptable. Do not add a new search provider or crawler just to duplicate an existing approved search path.

## 5. Components deliberately not added

- **Scrapy / Crawlee:** no general-purpose crawler framework is selected yet. Choose one only after inspecting actual target sites and proving that the source-specific MCP adapter plus HTTP/browser layer is insufficient.
- **Selenium:** not selected; do not add another browser automation stack alongside Playwright without a measured requirement.
- **PostgreSQL / Redis / Kubernetes:** not part of the local single-user stack.
- **A second scheduler or agent orchestration platform:** not required; Huginn already supplies these capabilities.

## 6. Recommended introduction order

1. Huginn source submodule (already pinned in the repository).
2. NVIDIA NIM (preferred initial model API).
3. Built-in Huginn web search/fetch where sufficient.
4. Playwright MCP only when a real source requires dynamic browser access.
5. One VALENTINA-specific MCP + SQLite service only when structured persistence/deduplication is justified.
6. UFO² only when a Windows desktop application workflow cannot be handled more simply.

Keep the number of maintained runtimes and services to the minimum that solves a demonstrated task.
