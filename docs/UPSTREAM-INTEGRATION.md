# Upstream source integration — Huginn

**Status:** Source linked in the VALENTINA repository as a pinned Git submodule  
**Upstream:** [scrypster/huginn](https://github.com/scrypster/huginn)  
**Release:** `v0.5.0`  
**Pinned commit:** `f01c1d84c08b1de972d8cc268da25602598adff8`  
**Local path:** `upstream/huginn`

## 1. What is included

The complete upstream Huginn Git repository is connected to VALENTINA as a Git submodule. After initialization, the local working tree contains the actual source code, documentation, embedded web UI source, Go modules/vendor tree, tests, and upstream build configuration under `upstream/huginn/`.

This is intentional. The audited upstream tree contains approximately 6,984 files, including 4,919 files under `vendor/`, with a combined file-size total of about 299 MB. Copying these files into the root VALENTINA repository would duplicate the upstream history and make every update and code review unnecessarily expensive.

A submodule gives us:
- The complete source tree locally for Cline, ChatGPT-connected coding tools and other agents.
- A precise, reviewable upstream commit recorded by the parent repository.
- A clean boundary between upstream Huginn and VALENTINA-specific code.
- A simple way to inspect upstream release changes and update the pinned revision intentionally.
- The original Huginn license and third-party notices inside the upstream source tree.

GitHub displays the submodule as a link to the pinned upstream commit. A normal clone must initialize submodules before agents can read or build the Huginn source.

## 2. Initialize the source

### New clone

Clone VALENTINA with submodules enabled:

```powershell
git clone --recurse-submodules https://github.com/MiguelDuval/VALENTINA.git
cd VALENTINA
`.\scripts\bootstrap.ps1
```

The bootstrap script initializes submodules and verifies the exact Huginn commit. On an existing clone, it is safe to run again.

### Existing clone

From the repository root:

```powershell
git submodule sync --recursive
git submodule update --init --recursive
git -C upstream/huginn rev-parse HEAD
```

Or run `.`scripts`/bootstrap.ps1` from PowerShell.

On Linux/macOS, use:

```bash
bash scripts/bootstrap.sh
```

Expected Huginn revision:

```text
f01c1d84c08b1de972d8cc268da25602598adff8
```

The source directory is not considered ready for implementation work until the submodule exists and its HEAD matches this pin.

## 3. Build Huginn from source on Windows

The upstream source identifies Go 1.25 or newer as its toolchain requirement. Its CI uses Node.js 20 for building the web frontend.

From PowerShell:

```powershell
Set-Location upstream/huginn
go version
node --version
npm --version

Push-Location web
npm ci
npm run build
Pop-Location

go build -tags embed_frontend -o huginn.exe .
```

This is the upstream build path adapted for a Windows executable name. It is documented here, not claimed as executed or validated on the owner's PC. Use the upstream README and release notes for any platform-specific changes.

Upstream references:
- [README / installation / build instructions](https://github.com/scrypster/huginn/blob/v0.5.0/README.md)
- [Go module and toolchain](https://github.com/scrypster/huginn/blob/v0.5.0/go.mod)
- [CI workflow](https://github.com/scrypster/huginn/blob/v0.5.0/.github/workflows/ci.yml)
- [Release v0.5.0](https://github.com/scrypster/huginn/releases/tag/v0.5.0)

## 4. Rules for coding agents

Every coding agent must:
1. Read `AGENTS.md`, `docs/ARCHITECTURE.md`, `docs/SECURITY-AND-SECRETS.md` and this document.
2. Initialize the submodule with `.`scripts`/bootstrap.ps1` on Windows or `bash scripts/bootstrap.sh` on Unix-like systems before inspecting or changing Huginn code.
3. Verify `git -C upstream/huginn rev-parse HEAD` equals the pin above.
4. Treat `upstream/huginn` as third-party upstream source, not as a directory for casual edits.
5. Put VALENTINA-specific adapters, source definitions, configuration templates and glue code outside the submodule.
6. Never commit local Huginn config, API tokens, credentials, databases, logs, browser profiles or operational data.
7. Never update the submodule pointer automatically. An upstream change requires a dedicated branch/PR, release/advisory review, an explicit commit change and updated documentation.

If a concrete requirement needs a Huginn code patch, first prove that configuration, an MCP adapter or a small external extension cannot solve it. Then propose an ADR and use a separate VALENTINA-owned fork if maintaining an upstream patch is genuinely necessary. Do not modify the public upstream repository.

## 5. Updating Huginn

Do not track a floating branch or blindly pull `develop`/ `main`.

For an approved upgrade:
1. Review the upstream release notes, security policy and license.
2. Choose an explicit release tag and resolve it to a commit SHA.
3. Update the submodule to that exact commit on a dedicated VALENTINA branch.
4. Re-run the bootstrap SHA check after changing the documented expected commit.
5. Review the upstream changelog and any local integration diffs.
6. Run the build and relevant tests where an environment is available.
7. Record the new release/tag/SHA in this document and open a PR.

Rollback is simply reverting the parent repository's submodule pointer in a reviewed commit; do not rewrite Git history.

## 6. Licensing

Huginn's `LICENSE` file is available at `upstream/huginn/LICENSE` and identifies the Business Source License 1.1 (BSL 1.1), with the Change Date and Change License stated in that file. This is source-available, not an OSI-approved open-source license.

The current Additional Use Grant permits the project's strictly internal use, provided Huginn or a derivative is not offered to third parties as a product or service. Preserve the upstream license and notices. A future change in distribution, audience, hosting, or commercial use requires a fresh license review before implementation.

