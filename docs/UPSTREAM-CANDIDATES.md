# Upstream project audit and adoption decision

**Audit date:** 2026-11-04  
**Status:** Huginn-first decision adopted for VALENTINA  
**Selected upstream:** [scrypster/huginn](https://github.com/scrypster/huginn)  
**Mandatory architecture:** [ARCHITECTURE.md](ARCHITECTURE.md)

## 1. Executive decision

Use **scrypster/huginn as the base application and main local control hub for VALENTINA**. Reuse its web UI, agents, sessions, model backends, MCP client, REST API, routines, workflows and scheduler. Do not build a replacement frontend, orchestration server or scheduler.

This is a reuse-first decision, not a claim that upstream Huginn already provides a complete lead-management or web-crawling product. The minimal custom work expected is:
- Configuration and VALENTINA-specific agents/workflows.
- Source-specific MCP adapters for websites that cannot be handled by existing read-only tools.
- A small local SQLite-backed VALENTINA MCP service only when a real source proof-of-concept demonstrates the need for structured deduplication and review history.
- An optional narrow UFO² bridge only for a proven GUI-only Windows workflow.

Do not fork or vendor the full upstream source for the initial build. Install a pinned official release and keep VALENTINA-specific files/configuration separate.

## 2. Resolve the name collision

There are two unrelated projects called Huginn:

| Project | Description | Decision |
|---|---|---|
| [scrypster/huginn](https://github.com/scrypster/huginn) | Go-based local AI-agent platform with web UI, multi-agent delegation, routines, workflows, MCP and model providers | **Selected** |
| [huginn/huginn](https://github.com/huginn/huginn) | Ruby-based event-monitoring and automation system | **Not selected; do not substitute** |

Do not use documentation or capabilities from the Ruby project to describe the selected Go project.

## 3. Audit snapshot of scrypster/huginn

Repository facts observed through GitHub on 2026-11-04:

| Item | Observed value |
|---|---|
| Repository | `scrypster/huginn` |
| Default branch | `develop` |
| Latest published release | `v0.5.0`, 2026-08-30 |
| Audited `develop` HEAD | `1e584648232118f06bce7d25f930850060fe698c` |
| HEAD commit date | 2026-08-30 |
| Last pushed timestamp shown by GitHub | 2026-08-30 |
| Implementation language | Go |
| Go toolchain in `go.mod` | Go 1.25+ |
| License | Business Source License 1.1 (BSL 1.1) |
| Upstream security policy | Latest release: full support; previous release: critical patches only; older releases: unsupported |

The GitHub repository metadata showed an update timestamp on 2026-11-03, but the code push timestamp and audited HEAD remained 2026-08-30. Do not mistake metadata activity for a new code release. The project is young and the release cadence/support capacity is a maintenance risk. Pin the tested release, monitor releases/advisories and validate upgrades before applying them.

Useful upstream references:
- [README and features](https://github.com/scrypster/huginn/blob/develop/README.md)
- [Releases](https://github.com/scrypster/huginn/releases)
- [Security policy](https://github.com/scrypster/huginn/blob/develop/SECURITY.md)
- [License](https://github.com/scrypster/huginn/blob/develop/LICENSE)
- [Workflows](https://github.com/scrypster/huginn/blob/develop/docs/features/workflows.md)
- [Routines and scheduler](https://github.com/scrypster/huginn/blob/develop/docs/features/routines.md)
- [MCP/local tools](https://github.com/scrypster/huginn/blob/develop/docs/features/local-tools.md)
- [Web UI](https://github.com/scrypster/huginn/blob/develop/docs/features/web-ui.md)
- [Configuration](https://github.com/scrypster/huginn/blob/develop/docs/reference/config.md)
- [Storage design](https://github.com/scrypster/huginn/blob/develop/docs/subsystems/storage.md)

## 4. What Huginn already gives VALENTINA

| Capability | Audit result | Reuse decision |
|---|---|---|
| Local web UI | Existing UI for chat, agents, inbox, sessions, connections, workflows, models and settings | Use as the initial VALENTINA interface |
| Agent system | Built-in and custom agents, tool allowlists, delegation | Configure dedicated discovery and analysis agents |
| Workflow engine | Sequential multi-step pipelines, step outputs, failure handling and run history | Use for search/parse/analyze/review pipelines |
| Routines | YAML-defined scheduled tasks | Reuse |
| Scheduler | In-process Go cron scheduler | Reuse; keep process running for scheduled checks |
| REST API | Authenticated API for sessions, agents, workflows and supporting resources; OpenAPI scaffold is incremental | Use documented routes only; verify each needed route |
| MCP | MCP client supports configured stdio/SSE servers and per-server approval gate | Preferred extension boundary for VALENTINA-specific tools |
| Model backends | Anthropic, OpenAI, Ollama, Google/Vertex and generic OpenAI-compatible `custom` backend in source | Configure NIM through `custom`, then test |
| Credentials | Supports environment-variable references and OS keyring references; keyring can be unavailable on some headless Linux systems | Prefer OS keyring or protected environment variables; verify actual storage behavior |
| Local state | Primarily local Pebble KV per workspace; some subsystems also use SQLite | Reuse for Huginn's own internal state only |

Huginn's web UI is a capable agent/workflow console, but it is not a ready-made commercial lead dashboard. Use it initially to avoid unnecessary interface development. A dedicated results screen is deferred until it proves necessary.

## 5. Important limitations for this use case

### 5.1 Not a general-purpose crawler

The built-in `web_search` tool calls the Brave Search API and is registered only when a Brave API key is configured. It returns a bounded set of search results (default five, maximum ten).

The built-in `fetch_url` tool performs an ordinary HTTP GET and converts HTML to Markdown. In the audited source it has a 10-second timeout and a 512 KB response cap, accepts HTTP/HTTPS and blocks private/internal destinations. It does not run JavaScript, operate a browser session or implement site-specific pagination/extraction.

Consequences:
- Configure a separate Brave Search credential if using built-in web search.
- Use ordinary fetch for simple public pages.
- Use a source-specific adapter for stable structured extraction.
- Add browser automation only for sources that require JavaScript or an interactive session and permit automated access.
- Do not describe Huginn alone as a universal web scraper.

### 5.2 Scheduler lifecycle and workflow semantics

Routines and Workflows are scheduled inside the Huginn process. If Huginn is stopped, scheduled tasks do not run; documentation says missed jobs are not automatically replayed after restart. The machine must be awake and Huginn must be running for continuous monitoring.

Workflows execute steps sequentially. Routines are AI-agent tasks that write output to the inbox; they are not a general-purpose event ingestion system. An unattended routine can stall if it attempts a write/exec tool and waits for approval. Keep discovery routines read-only and use visible failure/run history.

### 5.3 Storage is not a lead database

The primary Pebble KV store is designed for Huginn workspace state, code intelligence and internal records. It is local, single-machine and single-process. It is not a relational company/opportunity database with a normalized schema, source evidence and explicit review history.

When structured deduplication becomes necessary, create one compact local SQLite data store behind a VALENTINA MCP service. Do not add PostgreSQL, Redis or a database server for a single-user local tool.

### 5.4 Security posture

The server defaults to loopback and validates the web bind setting as localhost/127.0.0.1. The REST API uses a bearer token; the token endpoint is local and must not be exposed. MCP servers default to an approval gate unless explicitly disabled. Local tools can be allowlisted per agent.

Important risks:
- Never expose the local UI/API port to the LAN or Internet as part of the MVP.
- Do not put the API bearer token in workflow files, URLs, logs or source code.
- Avoid `--dangerously-skip-permissions` for normal or unattended discovery.
- Use a least-privilege discovery agent without shell, file-write or external-mutation tools.
- Treat website contents as untrusted input and validate custom connector networking to prevent SSRF.
- The source has a separate Brave API key requirement for web search.

## 6. NVIDIA NIM compatibility assessment

The audited Huginn source includes a generic `custom` backend that uses an OpenAI-compatible `/v1/chat/completions` endpoint and Bearer API-key authentication. Its configuration supports environment-variable references and keyring references.

NVIDIA NIM exposes an OpenAI-compatible hosted API. This makes the following configuration path plausible:

```json
{
  "backend": {
    "type": "external",
    "provider": "custom",
    "endpoint": "https://integrate.api.nvidia.com/v1",
    "api_key": "$NVIDIA_NIM_API_KEY"
  },
  "default_model": "<NIM_MODEL_ID>"
}
```

This is a proposed configuration example, not a verified live request. The NIM model ID must be selected from the user's available models. Before relying on NIM, run real tests for:
1. Basic chat completion.
2. Tool/function calling with the chosen model.
3. Streaming responses.
4. Vision input only if the chosen model and task require it.

If Huginn's current UI or config validation does not expose the `custom` provider correctly, prefer a documented configuration fix or a small adapter over changing the core backend. Keep the NIM key outside Git and never put it into prompts.

NVIDIA API reference: [NIM OpenAI-compatible API](https://docs.nvidia.com/nim/).

## 7. UFO² assessment

[Microsoft UFO²](https://github.com/microsoft/UFO/tree/main/ufo) is a Windows desktop AgentOS with HostAgent/AppAgents and hybrid UI Automation, Win32/WinCOM and visual controls. The UFO² README identifies the v2.0.0 release and MIT license at the repository root.

It is useful for difficult Windows GUI workflows, but:
- It is not a general website crawler.
- It is not needed for standard APIs, feeds or HTTP extraction.
- This audit did not establish a stable, ready-made Huginn-to-UFO² control API.
- A narrow local bridge would need a separate proof-of-concept and explicit permissions.

**Decision:** defer UFO² integration until a real source or desktop workflow proves that normal API/HTTP/MCP/browser methods are inadequate. Do not make it a required component.

## 8. License review

Huginn is under **Business Source License 1.1 (BSL 1.1)**. This is source-available, not an OSI-approved open-source license.

The checked-in Additional Use Grant permits production use for, among other cases, an individual or use that is strictly internal and not offered as a product or service to third parties in any form. That matches the current VALENTINA scope.

The license requires a separate commercial license for offering Huginn or a derivative as a hosted service, SaaS, managed platform, API service, commercial product, or paid/unpaid service to third parties. The stated change license is Apache 2.0 on 2026? No: the checked-in license states **March 7, 2030**. Preserve the original license and notices if source is copied or modified. Re-check the license before any change of distribution or audience.

## 9. Recommended minimal VALENTINA topology

```text
                User (local browser)
                        |
                        v
              HUGinn local web UI
              (primary VALENTINA UI)
                        |
               Agents / Workflows
                        |
            +-----------+-----------+
            |           |           |
            v           v           v
        Web search   HTTP fetch   NIM models
        (Brave)      (simple)    (analysis)
            |           |
            +-----+-----+
                  |
          Source-specific MCP
          adapter(s), only as needed
                  |
          Local SQLite registry
          (when structured storage
           and dedup are required)
                  |
             User review

       Optional later: UFO² Windows worker
       for a specific GUI-only workflow
```

The source adapter and SQLite registry should be one compact local MCP service, not a collection of independently deployed microservices. Initially defer that service until one representative source has been tested.

## 10. Rejected alternatives and why

| Alternative | Decision | Reason |
|---|---|---|
| Build a new React + FastAPI application first | Reject | Recreates UI, auth, workflows, model configuration and orchestration before proving a product-specific need |
| Use classic Ruby `huginn/huginn` instead | Reject | Different product; it is not the AI-agent platform selected by the owner |
| Put every responsibility into Huginn's Pebble store | Reject | Internal KV store does not provide the structured business-record model and provenance we need |
| Deploy PostgreSQL + Redis + Kubernetes | Reject for MVP | Unjustified operational and maintenance overhead for one local operator |
| Install several crawler frameworks together | Reject | Choose source-specific methods based on actual target sites |
| Make UFO² the primary web parser | Reject | GUI automation is more fragile and expensive than APIs/HTTP and does not solve structured storage |
| Fork Huginn immediately | Reject | Configuration and MCP integration should be tried first; a fork creates an ongoing upstream merge burden |

## 11. Go/no-go checklist before implementation

- [ ] Install the pinned upstream release on the intended local machine.
- [ ] Verify the web UI starts, restarts and persists state.
- [ ] Verify loopback binding and bearer-token behavior.
- [ ] Verify the Huginn `custom` backend against the chosen NIM endpoint.
- [ ] Run actual tool-call and streaming tests with the selected NIM model.
- [ ] Test one permitted real source end-to-end, including failures and rate limits.
- [ ] Confirm whether built-in web search is sufficient or a source connector is needed.
- [ ] Only then implement the minimum MCP + SQLite extension if the test demonstrates a structured-storage gap.
- [ ] Do not proceed to UFO² until one specific GUI-only workflow justifies it.

## 12. Audit limits

This was a source/documentation and GitHub metadata audit. It did not install or execute Huginn, did not send a live NVIDIA NIM request, did not test a real commercial source site, did not test UFO² orchestration, and did not verify runtime behavior on the owner's Windows machine. Those are explicit proof-of-concept tasks, not claimed accomplishments.
