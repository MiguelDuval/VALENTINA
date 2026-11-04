# VALENTINA

Internal-use project for a Valencia-focused commercial opportunity discovery and lead-research system.

## Purpose

VALENTINA is intended to collect and validate publicly available business opportunities relevant to a specific local business, reduce duplicates and irrelevant records, and present evidence-backed leads for human follow-up.

The system is for internal use. It is not intended to automate conversations with prospective customers or to be sold as a product.

**Repository visibility:** the GitHub repository is intentionally public. All committed code and Git history are public; credentials, real business/lead data, and private runtime configuration must remain outside Git. See the mandatory [coding-agent rules](AGENTS.md) and [security and secrets policy](docs/SECURITY-AND-SECRETS.md).

## Project status

**Bootstrap stage.** The repository has been created and the product requirements, target business, initial source landscape, architecture, and reusable upstream components have not yet been approved.

No application stack or upstream project has been selected in this repository yet.

## Working principles

- Preserve source URLs, timestamps, and evidence for every discovered opportunity.
- Separate companies, contacts, and individual opportunities.
- Deduplicate without collapsing genuinely different opportunities from the same company.
- Keep human review in the loop before commercial outreach.
- Respect source terms, privacy, and applicable Spanish and EU law.
- Prefer documented APIs and maintainable integrations over fragile GUI automation.
- Record upstream licenses and notices before reusing or modifying third-party code.

## Planned project documentation

- Product requirements and target-business profile
- Architecture decision records
- Upstream component and license evaluation
- Data model and source registry
- [Mandatory coding-agent rules](AGENTS.md)
- [Security, secrets, and configuration policy](docs/SECURITY-AND-SECRETS.md)
- Product privacy and operational policies
