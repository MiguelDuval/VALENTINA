# Mandatory instructions for coding agents

These instructions apply to every human or AI coding agent working in this repository, including ChatGPT, Cline, Copilot, and any future agent.

## Required reading

Before planning or editing code, read:

1. This file: `AGENTS.md`.
2. [Security, secrets, and configuration policy](docs/SECURITY-AND-SECRETS.md).
3. The relevant architecture/product documents for the task.

If a task conflicts with these rules, stop before making the conflicting change and explain the conflict. Do not silently override this policy.

## Repository visibility is intentional

- `MiguelDuval/VALENTINA` is intentionally **PUBLIC**. Do not change repository visibility, recommend a paid GitHub plan as a prerequisite, or create a private mirror unless the owner explicitly instructs you to.
- Public visibility means every committed source file, document, workflow, issue/PR content, and retained Git history may be read by anyone. The project's internal-use purpose does not make its public repository confidential.
- Never commit information that must remain confidential. Keep credentials, private operational data, customer/contact records, real leads, local databases, exports, and production data outside Git.
- Do not put secrets in code, tests, fixtures, documentation examples, screenshots, logs, CI artifacts, caches, PR descriptions, or issue comments.

## Credentials and NVIDIA NIM

- Follow [docs/SECURITY-AND-SECRETS.md](docs/SECURITY-AND-SECRETS.md) as a mandatory security requirement, not optional advice.
- During the initial development phase, NVIDIA NIM hosted inference is the preferred/initial model API. This is a phase-one choice, not a permanent architectural lock-in.
- The expected GitHub Actions repository-secret name for the initial NIM integration is `NVIDIA_NIM_API_KEY`. The actual key must be added by the repository owner under **Settings → Secrets and variables → Actions → Repository secrets**. Never ask the owner to paste the real key into chat, a commit, an issue, or a workflow file.
- Do not assume this secret exists. If it is unavailable, keep unit tests and offline development working with mocks; skip only the live NIM integration test and report that it was skipped because the secret was unavailable. Never fabricate a successful live API test.
- Local application users must configure their own API credentials through the application's settings or documented local environment configuration. Store saved credentials in the operating system's credential store where available; never bake them into an executable, APK, frontend bundle, source file, or default configuration.
- Keep the model-provider interface replaceable so NVIDIA NIM can be changed later without rewriting the core business logic.

## GitHub Actions and CI safety

- GitHub Actions secrets are for trusted CI workflows only. A secret is not safe merely because GitHub masks its value in logs: workflow code and third-party actions that receive it can still exfiltrate it.
- Expose a secret only to the exact job/step that requires it, using an environment variable. Do not echo it, print environment/context dumps, enable shell tracing around it, place it in command-line arguments, or upload it in artifacts/caches.
- Use least-privilege `permissions` for `GITHUB_TOKEN` and third-party actions. Pin third-party actions to full commit SHAs where practical.
- Do not expose secrets to untrusted pull-request code. Do not use `pull_request_target` to check out or execute untrusted PR code in a privileged context.
- Never construct shell commands directly from untrusted issue, PR, branch, commit, or scraped-page text. Validate and safely pass data.
- Use separate, low-privilege test credentials. Never use a personal or production API key in routine CI.

## Required Git workflow

- Work on a task-specific branch created from the current intended base. Do not commit directly to `main) unless the owner explicitly instructs you.
- Do not reset, rebase, force-push, rewrite history, or delete branches unless the owner explicitly authorizes the exact operation.
- Before editing, verify the repository is `MiguelDuval/VALENTINA`, report/check the branch and HEAD, and inspect the current working tree.
- Before committing, review the full diff for accidental credentials, personal data, generated files, and unrelated changes. Run relevant tests and report exactly what ran and what did not.
- Use a pull request to propose changes to `main`. Do not claim a PR was merged, CI passed, a key exists, or a test hit NVIDIA NIM unless the corresponding state was actually verified.
- Do not weaken or remove CI, security checks, or repository rules to make a failing change appear successful.

## Documentation and handoff

- Keep this file and the security policy aligned with the actual implementation. Update the documentation whenever credential flows, CI secret usage, or data-storage design changes.
- In every handoff, state the branch, commit SHA, PR (if created), test/CI evidence, skipped checks, and any secret or external-service prerequisites. Never include secret values.
