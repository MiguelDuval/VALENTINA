# VALENTINA UI — Implementation status

## Included
- VALENTINA dark palette with compatibility aliases for existing Tailwind classes.
- Sans-serif application typography; monospace retained for code.
- VALENTINA browser title, Russian document language, favicon and shell identity.
- Labeled main sidebar with compact icon-rail fallback at narrower widths.
- Russian primary navigation; Overview is the default route.
- Overview uses existing live APIs for health, agents, workflows and pending notifications.
- Honest commercial-opportunity empty state; no invented leads or statistics.
- Windows build script verifies upstream pin, applies first-party patch to a disposable copy, runs frontend tests and builds the embedded Go executable.
- A successful Windows Actions run uploads `VALENTINA.exe` as a SHA-labelled workflow artifact for 30 days; GitHub packages the download as a ZIP containing the executable.
- No GitHub Releases, release tags or public installer releases are used; test binaries are obtained only from completed successful Actions runs.

## Not complete or not verified
- No opportunity records, source connectors or CRM persistence.
- NVIDIA NIM chat, function calling and streaming are not certified by visual work.
- Not all legacy views translated or redesigned.
- Playwright visual regression and Windows runtime/browser interaction tests remain.
- NSIS installer remains a later milestone.
- The patch has not yet been run through a real build in this authoring environment; CI must verify it.

## Local Windows build
After initializing the submodule, run from repository root:
powershell -ExecutionPolicy Bypass -File .\scripts\build-valentina.ps1

The script fails closed on a pin mismatch or patch-apply failure. It never edits upstream/huginn. The generated `dist/VALENTINA.exe` is ignored by Git. The CI workflow uploads only this executable after all build checks pass; failed runs produce no artifact. Download it from the successful run's **Artifacts** section and extract the ZIP.
