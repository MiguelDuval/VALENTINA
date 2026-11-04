# Security, secrets, and configuration policy

**Status:** Mandatory project policy  
**Applies to:** All contributors, coding agents, workflows, tests, local installations, and deployment environments  
**Repository:** [MiguelDuval/VALENTINA](https://github.com/MiguelDuval/VALENTINA)  
**Last updated:** 2026-11-04

This policy is binding. It is not optional guidance. The root [`AGENTS.md`](../AGENTS.md) requires every coding agent to read and follow it.

## 1. Public repository decision

VALENTINA intentionally remains a **public GitHub repository**. This is a deliberate project constraint so development can use the public-repository capabilities available on the owner's GitHub Free account without making a paid plan a prerequisite.

This decision has an important consequence: **anything committed to this repository is public**. That includes application source code, documentation, workflow definitions, tests, issues and pull requests, and Git history. Deleting a file later does not guarantee that its contents disappear from clones, caches, forks, or historical commits.

Therefore:

- Do not commit passwords, API keys, access tokens, private certificates, private keys, confidential configuration, or other credentials.
- Do not commit customer/contact personal data, actual commercial leads, private business records, production databases, exports, backups, private logs, or screenshots containing sensitive information.
- Do not commit proprietary material or any source/data that the owner expects to remain confidential. The application's intended internal use does not make the public repository private.
- Do not change the repository to private as an autonomous action. This policy can be changed only by an explicit instruction from the repository owner.

The application may be for internal use while its development repository is public. Keep the operational data and runtime configuration separate from Git.

## 2. Where credentials belong

Use the following separation.

| Context | Approved location | Rules |
|---|---|---|
| GitHub Actions tests | GitHub **Actions repository secrets** (or narrowly scoped environment secrets when environments are introduced) | Only trusted workflows/jobs/steps may receive them. Never commit values. |
| Local development | Local environment variables or an ignored local `.env` file | Keep the file outside version control; use `.env.example` only for empty placeholders and non-secret defaults. |
| Installed VALENTINA application | Application settings backed by the operating system credential store/keychain where available | Let the local operator enter their own credentials. Do not store them in ordinary application tables, logs, or plaintext config. |
| Hosted/server deployment, if introduced | Server-side secret manager or protected deployment environment | Keep provider credentials on the server; do not expose them to browser/frontend clients. |
| Tests and fixtures | Mock credentials and fake data | No real API credentials or real customer/lead data in fixtures. |

The repository's `.gitignore` must exclude local environment files, credentials, local databases, logs, exports, and backups. A future `.env.example` may be committed only when it contains placeholders and no valid secrets.

A local credential store reduces accidental disclosure but is not an absolute security boundary against malware or an administrator with control of the machine. Do not describe locally stored credentials as impossible to extract.

## 3. NVIDIA NIM: initial development provider

For the initial development phase, NVIDIA NIM hosted inference is the preferred starting model API. This is an initial implementation choice, **not** a permanent provider lock-in or a requirement that prevents offline tests.

- The expected GitHub Actions secret name is `NVIDIA_NIM_API_KEY`.
- The owner configures the secret in **Settings → Secrets and variables → Actions → Repository secrets**. Do not put the key in source files, workflow YAML, repository variables, commits, documentation, chat messages, issues, or pull requests.
- A NIM endpoint/base URL and model identifier are configuration, not credentials. Keep them configurable; do not confuse a public endpoint URL with the API key.
- The key must be injected only into the specific trusted integration-test step that needs it.
- Local users configure their own NIM credentials in VALENTINA settings or the documented local secret mechanism. Do not reuse the owner's CI key as a distributed default.
- Keep model-provider access behind an adapter/interface so a different provider or local model can be introduced later without coupling core opportunity discovery logic to NIM-specific APIs.
- If `NVIDIA_NIM_API_KEY` is missing, unit tests and normal offline tests must still run using mocks. Mark only the live NIM integration test as skipped/not configured and report that accurately. Never claim a live provider test passed without a real request and evidence.
- Use a dedicated, low-privilege test key where the provider supports it. Monitor usage and revoke/rotate the key if exposure is suspected.

## 4. GitHub Actions secrets in a public repository

GitHub supports Actions secrets in public repositories. Their values are not shown to repository visitors, but **secrets are made available to workflow code when a workflow explicitly references them**. Any malicious or compromised workflow/action receiving a secret may be able to send it elsewhere. Log masking is a defense-in-depth feature, not a guarantee that a secret cannot leak.

Required controls:

1. Store secret values in **Settings → Secrets and variables → Actions → Repository secrets**. Use GitHub Actions variables only for non-sensitive settings such as a model name or base URL.
2. Use a secret only in the narrowest job/step that needs it. Prefer passing it as an environment variable, not embedding it in a command string or command-line argument.
3. Never use `echo`, debug dumps, `set -x`, verbose request tracing, or context dumps on secret-bearing steps. Never upload secret-bearing environment files, request traces, test output, caches, or artifacts.
4. Set explicit, minimal workflow `permissions` for `GITHUB_TOKEN`; grant write access only where a job demonstrably requires it.
5. Prefer trusted branch pushes or maintainer-triggered workflows for live tests that require credentials. Fork-originated pull-request workflows must not receive secrets.
6. Do not use `pull_request_target` to check out, build, or execute untrusted pull-request code while privileged tokens or secrets are available.
7. Treat workflow files and third-party Actions as executable code. Review changes to workflows and pin third-party Actions to full commit SHAs where practical.
8. Never interpolate untrusted PR/issue/comment/scraped content directly into shell commands. Validate inputs and pass data safely.
9. Do not use a real production key for ordinary CI. Use a separate test credential with the minimum permissions and quota necessary.

A secret being absent in a fork-based PR workflow is expected GitHub behavior, not a reason to weaken the workflow. Tests that need secrets should have a safe skip path and an explicit report.

## 5. Local application credential handling

When the application gains a settings interface:

- Provide a clear provider-configuration screen for the local operator to enter, replace, test, and remove their own provider key.
- Save secrets in the OS credential manager/keychain where supported. If a platform has no secure store, document the limitation and use the safest available protected storage; do not silently fall back to plaintext in the main database.
- Mask credentials in the UI after saving. Never return the full saved value to logs or diagnostics.
- Never compile a shared NIM key into an APK, EXE, installer, frontend bundle, source map, JavaScript bundle, or default configuration. A credential shipped to a client can be extracted.
- For any future browser-based or multi-user deployment, route provider requests through a trusted backend that keeps provider keys server-side; never expose a shared secret to browser code.
- Keep operational databases, real lead records, contact details, and backups outside the public repository.

## 6. Safe examples and configuration files

- Use obvious placeholders such as `replace-me`, `YOUR_NVIDIA_NIM_API_KEY`, or an empty value. They must never be valid credentials.
- If a sample config includes a key field, document the corresponding GitHub Secret/local setting without including a value.
- Keep `.env`, `.env.*`, local secret directories, private key/certificate files, database files, logs, exports, and backups ignored by Git.
- Review generated files, screenshots, crash dumps, and test reports before adding them. They may contain request headers, environment variables, user paths, or personal data.

## 7. Exposure or accidental commit response

If a credential may have been exposed:

1. Revoke or rotate it immediately at the provider; deleting the current file is not sufficient.
2. Identify affected workflows, artifacts, logs, caches, and deployments; disable or rotate impacted credentials as needed.
3. Remove the credential from tracked files. Rewrite Git history only under an explicit, coordinated owner-approved recovery procedure; history rewriting is not an autonomous agent action.
4. Review provider usage and access logs where available.
5. Record the incident without reproducing the credential.

Never paste the leaked value into an issue, chat, PR, or incident report.

## 8. Mandatory agent checklist

Before every commit, the contributor/agent must verify:

- [ ] No API keys, tokens, passwords, private keys, certificates, or real secrets are in the diff.
- [ ] No personal contact records, actual leads, production data, local databases, logs, exports, backups, or sensitive screenshots are in the diff.
- [ ] Local secret/config/data paths are covered by `.gitignore`.
- [ ] Any workflow that consumes a secret is trusted, narrowly scoped, least-privileged, and safe for untrusted PRs.
- [ ] NIM integration tests use `NVIDIA_NIM_API_KEY` only when configured; otherwise the live test is clearly skipped and mocks cover offline testing.
- [ ] Tests and CI results are reported truthfully; no skipped test is described as passing.
- [ ] Documentation matches the implementation.

**No exception may be made by an autonomous coding agent.** Any policy change requires an explicit decision by the repository owner, followed by an update to this document and `AGENTS.md`.
