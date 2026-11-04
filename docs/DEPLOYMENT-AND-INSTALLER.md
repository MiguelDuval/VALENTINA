# VALENTINA — deployment and Windows installer decision

**Status:** Accepted direction under ADR-002, refined by owner-approved ADR-005  
**Date:** 2026-11-04  
**Initial target:** Windows 11 x64, one local operator  
**User-facing package:** `VALENTINA-Setup-x64.exe`

## 1. Decision

Deliver one per-user Windows installer containing the pinned Go application with the VALENTINA-branded Vue web client embedded in its executable.

The normal launch flow is:

1. User launches **VALENTINA** from the Start Menu.
2. The local Go process starts and binds only to `127.0.0.1`.
3. The default browser opens the VALENTINA Overview / Review Queue.
4. The same local application serves the VALENTINA domain routes and the retained AI/agent/workflow administration screens.

Do not package a separate Go companion just to provide another CRM website. A second process requires a demonstrated blocker and a new ADR. Keep the local browser application and domain functionality inside the existing app where practical.

## 2. Runtime and build inputs

The pinned upstream project contains:

- A Go application (the server/runtime).
- A Vue 3 / TypeScript / Vite / Tailwind browser frontend under `web/`.
- An existing local static-file server.
- An `embed_frontend` Go build path that embeds `internal/server/dist` into the executable.

Build-time requirements are Go 1.25+ and the Node.js/npm toolchain required by the pinned frontend. These are build tools, not end-user prerequisites.

The actual build pipeline must:

1. Initialize and verify the exact submodule SHA in `docs/UPSTREAM-INTEGRATION.md`.
2. Create a clean temporary worktree/copy of the pinned upstream.
3. Apply the tracked VALENTINA UI patch set from outside the submodule.
4. Build the Vue frontend into `internal/server/dist` using the pinned lockfile/toolchain.
5. Compile the Go application with `-tags embed_frontend`.
6. Fail if patches do not apply, dependencies differ unexpectedly, tests fail or the embedded frontend is missing.

Do not modify the submodule pointer or commit ad-hoc submodule working-tree edits for routine branding changes.

## 3. Package structure

One Setup executable installs:

- The single Go application executable with the VALENTINA browser UI embedded.
- Required local defaults and safe templates.
- License files and third-party notices, including the complete pinned upstream BSL 1.1 license.
- Start Menu shortcut named **VALENTINA**, uninstaller and version metadata.

Use NSIS with a pinned version. Do not bundle Go, Node.js, npm, Python, Docker, WSL, PostgreSQL, Redis, Kubernetes, a local LLM, Playwright browsers or UFO² in the base installation.

Do not put API keys, personal profiles, real leads, local databases or production data in the installer.

## 4. First-run and NVIDIA NIM

The first-run UI belongs to VALENTINA and should be short, Russian-language and understandable without technical knowledge.

It must offer:

- NVIDIA NIM as the preferred initial provider.
- Endpoint default `https://integrate.api.nvidia.com/v1`.
- An editable model ID from the models available to the owner's account.
- A masked API-key input saved only through Windows Credential Manager or another reviewed protected local credential mechanism.
- A Test Connection action and a clear non-secret result.
- An option to configure the business profile locally.

Never store credentials in frontend bundles, localStorage, URL query strings, command-line arguments, registry values, logs, screenshots, installer resources or Git. Never claim the key has been validated until a real request succeeds.

## 5. Local-only security and mutable data

- Bind the application to `127.0.0.1`; do not bind to `0.0.0.0`.
- Do not open inbound firewall ports or expose a reverse proxy.
- Disable cloud relay, remote access and collaboration features in the local product unless a new security review authorizes them.
- Keep application binaries separate from mutable user data.
- Keep owner profile, source registry, SQLite database, evidence cache, config, credentials and backups outside the install directory and outside public Git.
- Preserve user data during upgrade and uninstall by default.
- Start-at-login is explicit opt-in; do not install a privileged Windows service by default.
- Logs must be useful but must not contain secrets, full private profile contents or unnecessary personal data.

## 6. UI branding and license display

The application must present VALENTINA as its normal identity: name, logo, browser title, favicon, start page, navigation, welcome flow and ordinary messages. No upstream product branding is needed in the ordinary interface.

The complete required upstream license and copyright notices must remain present. Provide a clearly reachable **About → Licenses** screen containing the full BSL 1.1 license text and preserve the original license file in the installed application. This legal screen is not a second product brand or a normal navigation headline.

The owner's declared use is local, single-operator and internal, with no hosting or service offering to third parties. Any change to distribution, hosting, external users or commercial packaging requires a fresh review of the pinned license and its Additional Use Grant before release.

## 7. Windows acceptance gates

A build is not a successful installation. Verify on a real Windows 11 x64 machine:

1. Clean per-user installation without administrator rights.
2. Start Menu launch and automatic opening of the VALENTINA browser UI.
3. Loopback-only binding and no remote/cloud connectivity.
4. NIM first-run configuration and real connection test.
5. Manual execution of one permitted source workflow.
6. Local persistence after application restart.
7. Backup and restore in a clean local profile.
8. Upgrade without loss of settings, credentials or business data.
9. Uninstall that preserves user data by default.
10. Presence of the complete license texts and third-party notices.

The first private build may be unsigned and may trigger Windows SmartScreen. Paid code signing and public distribution are deferred unless they become an explicit requirement.
