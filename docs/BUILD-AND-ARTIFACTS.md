# VALENTINA — Build and Artifacts

**Status:** Draft  
**Date:** 2026-11-07

This document describes how a single, reproducible VALENTINA Windows x64 artifact is built, what it contains, and how an owner can verify it.

## 1. What a VALENTINA build contains

One VALENTINA build produces a single GitHub Actions artifact ZIP named:

```
VALENTINA-windows-x64-<FULL_COMMIT_SHA>.zip
```

Inside the ZIP:

| File | Purpose |
|------|---------|
| `VALENTINA.exe` | The pinned Huginn Go runtime with the VALENTINA frontend embedded (`-tags embed_frontend`). |
| `VALENTINA-BUILD-MANIFEST.json` | Machine-readable build manifest (see schema below). |
| `SHA256SUMS.txt` | `SHA256  VALENTINA.exe` line for independent checksum verification. |
| `README.txt` | Human-readable build identification and verification instructions. |
| `LICENSES/` | Full upstream license and notice texts (BSL 1.1 and others), not truncated. |

**Not included:** SQLite database, real CRM records, leads, user profiles, Huginn configuration, API keys, OAuth tokens, browser profiles, secrets, temporary build files, `node_modules`, Go compiler, Node.js runtime, Python runtime, Git repositories.

## 2. Staging structure

The build script (`scripts/build-valentina.ps1`) does **not** compile in the VALENTINA repository root. It:

1. Clones the pinned upstream Huginn submodule (`f01c1d84c08b1de972d8cc268da25602598adff8`) into a temporary `source/` directory.
2. Applies all tracked first-party patches (`patches/upstream-ui/*.patch`) with `git apply --check` (fail closed if context differs).
3. Copies VALENTINA-first-party Go source into `source/internal/valentina/internal/{domain,api}` — matching the import path `github.com/scrypster/huginn/internal/valentina/internal/...`.
4. Copies VALENTINA-first-party Vue views into `source/web/src/views/`.
5. Verifies that the upstream server imports the CRM API package.
6. Runs `go test ./internal/valentina/internal/...` in the upstream module context.
7. Runs `npm ci`, `npx vue-tsc --noEmit`, `npm test`, `npm run build` in `source/web/`.
8. Compiles `go build -tags embed_frontend -o VALENTINA.exe` from `source/`.
9. Generates `VALENTINA-BUILD-MANIFEST.json`, computes `SHA256SUMS.txt`, assembles the package directory, and zips it.

## 3. Build manifest

File: `VALENTINA-BUILD-MANIFEST.json`  
Schema: `schemas/valentina-build-manifest.schema.json`

The manifest is generated **after** the EXE is built and all validation checks pass. It contains:

- `schema_version` — manifest schema version (1).
- `product_name` / `product_version`.
- `source` — repository, branch, full commit SHA, parent commit SHA, dirty working tree flag, build timestamp UTC.
- `upstream` — repository, exact pinned SHA, verified flag.
- `build_environment` — OS, architecture, Go/Node/npm versions, workflow name, GitHub Actions run ID/attempt/URL.
- `validation` — status/command/result for each mandatory check (Go tests, frontend tests, Vue typecheck, production build, EXE compilation, patch application, upstream pin verification).
- `binary` — file name, size, SHA256, platform, architecture.
- `security` — guarantees that no secrets, database, user profile, scraping, or credentials are embedded.
- `artifact` — artifact name, SHA256 digest, retention period, GitHub Actions run URL.

**No secrets, environment variable values, user data, or local database contents are ever included.**

## 4. Checksum verification

After downloading the artifact ZIP and extracting it:

```powershell
certutil -hashfile VALENTINA.exe SHA256
```

Compare the output with:
- The `sha256` field in `VALENTINA-BUILD-MANIFEST.json` → `binary.sha256`.
- The line in `SHA256SUMS.txt`.

The SHA256 of the **ZIP archive itself** is reported separately in the manifest (`artifact.sha256_digest`) and in the GitHub Actions Job Summary; it is never stored inside the archive it describes.

## 5. Where to download

Artifacts are uploaded only to successful GitHub Actions workflow runs for the `feature/valentina-dual-pipeline-crm-20261106` branch or PR #5. They are retained for 30 days.

Download location (example):

```
https://github.com/MiguelDuval/VALENTINA/actions/runs/<RUN_ID>
```

## 6. CI workflow

The existing `.github/workflows/valentina-ui-build.yml` runs on push to `feature/**`, on PR, and on `workflow_dispatch`. It:

- Checks out the exact SHA with submodules.
- Verifies the upstream Huginn submodule pin.
- Applies all first-party patches.
- Runs Go tests, server integration tests, frontend tests, Vue typecheck, and production build.
- Compiles the Windows executable.
- Generates and validates the build manifest.
- Assembles and uploads the single artifact ZIP.

A failed or cancelled build never exposes an artifact. Mandatory tests cannot be skipped via `continue-on-error`.

## 7. Security

- No unattended scraping; source-policy adapter not implemented; source status remains `pending_review`.
- Loopback-only application; existing Huginn authentication preserved.
- No public webhooks; no external messaging.
- NVIDIA NIM not connected without credentials.
- No secrets embedded in the executable or manifest.

## 8. Related documents

- `docs/ADR-006-DUAL-PIPELINE-CRM-INTAKE.md` — CRM integration decision.
- `docs/MVP-DELIVERY-ROADMAP.md` — MVP delivery stages (Stage 6: Windows installer).
- `docs/SECURITY-AND-SECRETS.md` — credential and secret policy.
- `docs/DEPLOYMENT-AND-INSTALLER.md` — deployment and installer policy.
