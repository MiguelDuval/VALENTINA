# VALENTINA — deployment and installer architecture

**Decision:** ADR-002 — Windows-first single-user installation package  
**Status:** Accepted architecture direction; implementation is deferred until the pinned Huginn runtime has been built and its local startup path verified  
**Date:** 2026-11-04  
**Applies to:** packaging, installation, first-run setup, upgrades, runtime dependencies and local administration

## 1. Decision summary

VALENTINA should have a simple, self-contained installer for the owner's Windows computer. The intended first deliverable is:

- A versioned Windows x64 installer named like `VALENTINA-Setup-x64.exe`.
- The pinned Huginn Go application with its web frontend embedded, plus one small VALENTINA-owned Go companion for the domain UI/data/source-policy boundary (exact process boundary subject to a dedicated implementation ADR).
- A Start Menu shortcut named **VALENTINA** that starts the VALENTINA interface and the required local components. Huginn's advanced admin UI remains separately accessible.
- A guided first-run configuration for the selected model provider and local application settings, without embedding credentials in the installer or executable.
- An uninstall/upgrade lifecycle that preserves operational data unless the owner explicitly chooses to delete it.

The installer is a delivery mechanism. It is not a reason to create a second backend, a new frontend framework, a local database server, or a container stack.

**Do not attempt to create one universal binary containing every optional environment.** Keep the normal installation small and add optional runtimes only when the corresponding integration is enabled.

## 2. Why this is feasible with the selected foundation

The pinned Huginn release is a Go application. Its upstream build process embeds the web frontend into the executable using the `embed_frontend` build tag. The upstream repository documents a standalone Windows executable and a single-binary installation model.

Build-time tools are different from end-user runtime requirements:

| Component | Build-time role | Required on the owner's PC for normal VALENTINA use? |
|---|---|---|
| Go 1.25+ | Compile Huginn and the VALENTINA Go companion | No |
| Node.js 20 + npm | Build Huginn's web frontend assets | No; assets are embedded into the executable |
| NSIS | Assemble the Windows installer | No |
| Web browser | Display the local web UI | Use the browser already installed on Windows |
| NVIDIA NIM | Hosted model inference through HTTPS | No local model runtime or local GPU required for this selected mode |
| SQLite / Huginn storage | Local application data, embedded in the application stack | No separate database server |
| Playwright + browser | Optional dynamic web-page automation | No, unless enabled for an approved source |
| Python / UFO² | Optional Windows GUI automation | No, unless a proven GUI-only workflow requires it |

The exact runtime dependency list must be verified against the pinned source and Windows build before the first installer is declared production-ready. Do not silently add a separate runtime because it was used on the build machine.

## 3. First target and packaging format

### 3.1 First target

- **Initial target:** Windows 11 x64, matching the current owner workstation.
- **Later targets:** Linux and macOS may receive their own native packages after the Windows deployment path is stable.
- Do not promise one identical installer for all operating systems. Each platform has its own executable format, paths, startup model, credential store and signing requirements.

### 3.2 Package structure

Use one installer package containing the two application components:

1. **Huginn executable:** the pinned Huginn Go program, built with the frontend embedded. Keep upstream behavior and license notices intact.
2. **VALENTINA companion executable:** a small Go application that owns the local domain HTTP UI/API, SQLite records, source-policy enforcement and MCP integration. It should use embedded server-rendered assets/templates unless a reviewed ADR justifies a frontend build framework.
3. **Installer executable:** one compressed Windows Setup program that installs both components, establishes predictable startup/shutdown behavior, creates the VALENTINA Start Menu entry and provides upgrade/uninstall behavior.

**Preferred Windows installer technology: NSIS.** It is an established, scriptable, open-source installer system that can create self-contained Setup executables, shortcuts, uninstallers and optional installation components. Pin the exact NSIS version in the build process. Do not add an installer framework to the running application.

Official references:
- [NSIS project](https://nsis.sourceforge.io/Main_Page)
- [NSIS features](https://nsis.sourceforge.io/Features)
- [Huginn v0.5.0 source and build instructions](https://github.com/scrypster/huginn/tree/v0.5.0)

Inno Setup is a technically viable alternative, but NSIS is preferred for this project to avoid an additional commercial-use licensing question in the installer toolchain.

## 4. What the normal installer must and must not include

### Include

- The compiled Huginn application with its web frontend embedded.
- The compiled VALENTINA companion executable and its embedded UI assets.
- The minimum VALENTINA-owned defaults and templates required for first launch.
- Required license texts and third-party notices, including the Huginn BSL 1.1 license.
- Start Menu shortcut(s), uninstall entry and version metadata.
- A first-run path to configure the model endpoint, model ID and credentials.
- A local health/startup check with a clear error message if the server cannot start.

### Do not bundle by default

- Go SDK, Node.js, npm, Python or a package manager.
- Docker, WSL, PostgreSQL, Redis, Kubernetes or a separate database server.
- A local LLM, Ollama model download or NVIDIA NIM container. The selected initial model path is NVIDIA NIM hosted inference.
- Playwright, Chromium/Firefox/WebKit downloads or an always-running browser automation service.
- UFO² and its Python environment.
- Real API keys, bearer tokens, user data, sample lead records or a preconfigured production database.

The normal user installation must not require the owner to open a terminal, clone Git repositories, initialize submodules, install Go/Node/Python, or run build commands.

## 5. First-run experience

The intended user flow is:

1. Download and run `VALENTINA-Setup-x64.exe`.
2. Complete a short installer wizard; install for the current Windows user by default and do not require elevation unless a later, justified feature needs it.
3. Launch VALENTINA from the Start Menu.
4. Start the VALENTINA companion and Huginn local process in the defined order; open the VALENTINA CRM interface in the default browser.
5. Complete a first-run setup for NVIDIA NIM: endpoint, account-available model ID and API credential.
6. Validate configuration with a clear connection/status result. Do not claim NIM is configured until an actual request succeeds.
7. Keep the app bound to loopback only. The first installation must not expose the UI/API to the LAN or Internet.

Start-at-login must be an explicit opt-in, not a silent default. The initial release should not install a privileged Windows service or require administrator rights merely to keep the local app available.

## 6. Interface and administration boundaries

There are two different interface needs:

### 6.1 Application and system administration

Use the existing Huginn local web UI initially for:
- Agent and permission administration.
- Model/provider configuration.
- MCP connections and integrations.
- Routines, workflows, schedules and execution history.
- Application sessions and operational status.

Use Huginn's existing tray/web startup path where it works on Windows. The installer should launch and link to this existing application rather than build a second desktop shell.

### 6.2 VALENTINA CRM/domain screens

Huginn is an agent/workflow platform, not a complete domain-specific CRM. Its initial UI must not be represented as if it already provides:
- A dedicated company directory.
- Opportunity/lead cards and a pipeline.
- Business-specific statuses, tags and review queues.
- Evidence/source history per opportunity.
- Domain-specific deduplication and merge review.

The VALENTINA CRM/domain UI is the main user-facing workspace, not a future cosmetic enhancement. The recommended implementation is a small Go companion that serves a local interface and reads/writes the approved SQLite domain store while exposing narrowly scoped MCP tools to Huginn. Prefer server-rendered HTML and minimal JavaScript/HTMX; do not introduce a separate React/FastAPI stack by default. The exact process boundary, startup lifecycle and API contract require a dedicated owner-reviewed ADR before coding.

## 7. Data, credentials and uninstall safety

- Keep application binaries separate from mutable user data.
- Use Huginn's supported per-user data/config location unless source verification identifies a concrete need for a VALENTINA-owned data directory. Do not invent a path or move upstream data without an explicit migration.
- User data, local databases, logs, backups, real opportunities and source evidence must remain outside the public Git repository and outside the installer payload.
- Store the NIM key in the operating system credential store or another protected local secret mechanism. The installer, executable resources, default config, shortcuts, registry values and logs must never contain the actual key.
- Keep the Huginn web UI/API bound to `127.0.0.1` / `localhost`; do not open inbound firewall ports.
- An upgrade may replace application files but must preserve the user's configuration, secrets, workflows, history and business data.
- Uninstall must remove application files and shortcuts but preserve user data by default. Offer a separate, clearly worded data-removal choice only when the owner explicitly selects it.
- Back up user data before migrations that may alter its schema or format. Failed upgrades must leave a documented rollback path.

## 8. Optional integrations

Optional components remain optional and separate from the base installer:

- **Playwright MCP:** install or provision only after an approved source demonstrably requires browser rendering. Its Node.js and browser dependencies are not prerequisites for normal VALENTINA operation.
- **UFO²:** defer until a specific Windows GUI-only workflow has been justified. It may require a separate Python environment and must not enlarge the normal installer.
- **Brave Search:** configure its own API key only when that provider is enabled; it is independent of NVIDIA NIM.
- **VALENTINA Go companion + SQLite + MCP + local UI:** this is now the recommended minimum domain component for the CRM interface and hard source-policy enforcement. It is one small local component, not a separate database server or per-source microservice. Its implementation is still gated by the dedicated process-boundary/API ADR.

See [OPTIONAL-INTEGRATIONS.md](OPTIONAL-INTEGRATIONS.md).

## 9. Build and release requirements

When implementation begins:

1. Build the pinned Huginn release in a controlled Windows build environment using the exact Go and Node toolchain versions required by that release.
2. Build the Huginn web frontend, then compile it with `go build -tags embed_frontend`.
3. Build the VALENTINA companion with its UI assets embedded; validate MCP and HTTP startup/shutdown behavior.
4. Run upstream tests and the Windows build/smoke checks that are available.
7. Package both application components and required notices using a pinned NSIS version.
8. Verify installation, first launch, companion/Huginn startup and shutdown, loopback binding, browser opening, UI-to-SQLite persistence, upgrade, uninstall and user-data preservation on a real Windows 11 machine.
7. Record SHA-256 hashes and version metadata for the installer and both application executables.
8. Never claim an installation test succeeded based solely on a successful compilation.

Do not publish VALENTINA/Huginn binaries as public GitHub releases by default. The repository is public, but the pinned upstream Huginn license is BSL 1.1 with a limited internal-use grant. Keep initial installation artifacts for the owner's internal use. Any wider distribution, hosted offering, third-party delivery or commercial product packaging requires a fresh license review before publication.

The build must never receive or package real NIM/Brave keys. Live external-service tests use only narrowly scoped credentials and only in the exact job that needs them.

## 10. Deferred work — do not implement prematurely

This decision does not authorize:
- A new desktop shell (Wails, Tauri, Electron or custom WebView).
- A new independent admin frontend.
- A Windows service or always-on privileged background agent.
- An auto-updater.
- Bundling optional Node/Python/browser runtimes.
- Public distribution of the installer.
- Any modification or fork of upstream Huginn.

Implement the installer only after the pinned application starts correctly on the owner's Windows system and the actual runtime/data/credential paths are verified.

## 11. Acceptance criteria for the first installer

The first user-facing installer is acceptable only when:

- A clean Windows 11 x64 machine/user profile can install and launch the VALENTINA UI plus required local components without Go, Node.js, Python, Git, Docker or administrator-level setup.
- The local web UI opens from the Start Menu entry and is reachable only through loopback.
- The user can configure the NIM endpoint/model/key without exposing the key in files, logs or process arguments.
- Restarting Windows or the app does not lose user configuration or data.
- Upgrade preserves user configuration and data.
- Uninstall removes the application but preserves user data by default.
- The installer and application have verifiable version information and SHA-256 hashes.
- All required third-party notices and license files are included.
- No real leads, contacts, source data, credentials or local database are included in the package or public repository.
