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

The first private test build may be unsigned and may trigger Windows SmartScreen. Paid code signing is deferred unless it becomes necessary for local testing.

## 8. Test executable distribution — owner requirement

VALENTINA does not publish official releases. Do not create GitHub Releases, release tags or public installer-release pages.

For owner testing, a successful Windows GitHub Actions run attaches the built `VALENTINA.exe` as a SHA-labelled workflow artifact. The download is a ZIP containing the executable, retained for 30 days. Failed or cancelled runs have no downloadable executable artifact. The owner downloads it from the completed run's **Artifacts** section.

The workflow artifact is a test-build delivery mechanism, not a claim that the product or Windows 11 owner acceptance gates are complete. The standalone `VALENTINA.exe` artifact remains unchanged. The workflow now also builds and tests the NSIS setup before uploading either binary artifact.

## 9. Stage 6 implementation (2026-11-26)

`scripts/package-valentina-installer.ps1` requires PowerShell 7 and consumes the existing `dist/package`. It verifies the executable SHA256/size, AMD64 PE header, checksum file, build validation metadata, exact clean upstream pin and byte-identical upstream license. In CI the build manifest must identify the checked-out HEAD. Output must be a new directory: stale setups are never silently reused.

The compiler is **NSIS 3.11**, a reproducibility pin, not a claim of latest version:
- Verified download mirror: `https://distfiles.macports.org/nsis/nsis-3.11.zip` (original NSIS 3.11 ZIP; verified against pinned SHA256). The SourceForge redirect returned HTML on MIGUELHP, so packaging uses this mirror and still fails closed on any unexpected bytes.
- SHA256: `c7d27f780ddb6cffb4730138cd1591e841f4b7edb155856901cdf5f214394fa1`
- Download uses HTTPS, checks every redirect and verifies SHA256 before extracting/executing anything. A failed download or hash mismatch stops packaging; no package-manager installation, elevation or fallback version.
- Compiler scratch lives in unique ignored `build/installer-*` directories. Only the explicit payload is packed; the compiler and scratch never enter an artifact.
- Offline: `pwsh -NoProfile -File scripts/package-valentina-installer.ps1 -NsisArchivePath <verified-archive.zip>`. Optional `-MakeNsisPath <local-makensis.exe>` requires that archive and verifies every distributed tool file against it. Compilation uses a fresh verified extraction with `/NOCONFIG`, avoiding extra local plugins/configuration.

`installer/valentina.nsi` requests user privileges, defaults to `%LOCALAPPDATA%\Programs\VALENTINA`, creates one current-user Start Menu shortcut and 64-bit HKCU uninstall entry. It installs only EXE, README, build manifest, checksum, full upstream license, available third-party notices and its generated uninstaller. No service, firewall rule, startup task, model or runtime dependency is installed. The shortcut runs `VALENTINA.exe serve` with the user profile as its working directory. Without `serve`, pinned Huginn starts a terminal UI instead of its local web server. On a first run without configured agents, the daemon opens the browser; later starts may use the tray. Interactive Windows 11 acceptance remains outstanding.

Fresh targets must be empty, a different registered install location cannot be taken over, and junction/symlink target ancestry is rejected. Upgrade/uninstall checks executable write access without truncating it; a mapped or otherwise locked EXE fails with exit 2 and is never killed/replaced at reboot. Other guard failures return 1/3, write failures 4. Uninstall deletes only explicitly installed files, with nonrecursive directory removal; unknown files and all mutable application data are preserved. Concurrent installer instances in the session are rejected.

Safety contract and negative package tests run before packaging. `scripts/test-installer.ps1 -DisposableRunner` refuses ordinary workstations and self-hosted runners. On a disposable GitHub-hosted Windows runner it verifies fresh silent installation into a unique path with spaces, occupied/other-location rejection, exact EXE/license/shortcut/HKCU metadata, locked-file upgrade/uninstall rejection, installed and upgraded runtime HMAC/business/persistence/loopback checks, silent uninstall, and preservation of synthetic sentinels, actual retained runtime config/database files and unknown program-directory files. No fallback deletion hides failures. Existing runtime smoke behavior is unchanged unless the guarded retained-profile option is explicitly selected.

After all gates pass, `VALENTINA-installer-windows-x64-<GITSHA>` contains `VALENTINA-Setup-x64.exe`, `SHA256SUMS.txt` and `VALENTINA-INSTALLER-MANIFEST.json`, retained 30 days. The existing EXE artifact remains separate. No GitHub Release/tag is created. Tests do not use NIM secrets or live sources. GitHub Windows runner acceptance does **not** establish Windows 11 owner-machine acceptance, non-admin interactive UX, browser auto-opening, backup/restore or live provider/source readiness.

Legal scope: full pinned BSL is installed verbatim. Available tracked vendored license/notice text is collected into one notices file, preserving paths/text. This is **not a complete frontend/transitive license inventory**; no fonts are added by this installer. About → Licenses UI and complete dependency audit remain acceptance limitations; this packaging pass does not rebuild the frontend solely to add that screen.

### Кратко для владельца

1. Откройте успешный запуск Actions → Artifacts → `VALENTINA-installer-windows-x64-<SHA>`.
2. Распакуйте ZIP-обёртку. Сверьте SHA256 Setup с `SHA256SUMS.txt` (например, `Get-FileHash .\VALENTINA-Setup-x64.exe`).
3. Запустите Setup без прав администратора, затем VALENTINA из меню Пуск.
4. Для обновления сначала закройте приложение и установите новую сборку в тот же каталог.
5. Удаление через настройки Windows сохраняет данные и ключи. Установщик не предлагает удалять их.

Локальная компиляция на тестовом EXE не заменяет успешный запуск CI. Не распространяйте локальные fixture-сборки из `build/`.
