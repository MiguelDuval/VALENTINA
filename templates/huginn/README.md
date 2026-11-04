# Huginn starter kit for VALENTINA

These files are reviewed templates, not active application configuration. They are not automatically installed, enabled or scheduled.

## Windows local paths

Huginn uses the per-user .huginn directory. In PowerShell, the templates can be copied to these locations:

- Agents: %USERPROFILE%\.huginn\agents\
- Skills: %USERPROFILE%\.huginn\skills\
- Workflows: %USERPROFILE%\.huginn\workflows\

Copy the three JSON agent definitions from agents/ into the agents directory. Do not overwrite an existing agent with the same name without first reviewing it.

Copy the Markdown files from skills/ into the skills directory. Then open Huginn's web UI → Settings → Skills and enable the three VALENTINA skills. A skill file on disk is not automatically enabled.

Copy valentina-research-manual.yaml from workflows/ into the workflows directory. It is disabled and manual-only by default. Review it and supply the workflow trigger inputs before running.

## Important security limitation

The current Huginn built-in `fetch_url` tool does not technically enforce the Source Registry's allowed host/path list. Its SSRF protections and response limits are not equivalent to a VALENTINA source allowlist. The starter prompts are not a security boundary.

For this reason the workflow is disabled and manual-only. Do not enable unattended runs or treat the source registry as technically enforced until the dedicated VALENTINA fetch/search MCP adapter validates source IDs, hosts, paths, redirects and resolved IPs server-side.

## Agent safety defaults

- VALENTINA-Scout: web_search and fetch_url only. Brave Search requires its own separately configured key.
- VALENTINA-Verifier: fetch_url only.
- VALENTINA-Assessor: no local tools.
- All three have an empty external toolbelt and disabled long-term memory.
- None has shell, file-write, Git-mutation or outbound-action permissions.
- Do not run unattended discovery with --dangerously-skip-permissions.

## Prerequisites

1. Initialize the pinned Huginn submodule and build/install the application.
2. Configure the initial NIM backend using a model ID available to the owner's account.
3. Complete the real local owner profile by copying templates/owner-profile.template.yaml to private/owner-profile.yaml.
4. Review and approve at least one source profile before discovery.
5. Configure Brave Search separately if web_search is to be used.
6. Read PRODUCT-AND-WORKFLOWS.md, DATA-AND-EVIDENCE.md, SOURCE-REGISTRY.md and QUALITY-AND-OPERATIONS.md.

## Important limitations

This starter kit produces candidate and review outputs only. It does not persist opportunities into a domain database, does not send outreach and does not provide a purpose-built CRM screen. Those capabilities require a later, separately reviewed VALENTINA data/UI implementation.
