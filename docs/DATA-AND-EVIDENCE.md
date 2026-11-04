# VALENTINA data model and evidence contract

**Status:** Provisional logical data contract v0.1; not a database migration  
**Date:** 2026-11-04

## 1. Data ownership

Huginn remains responsible for application state, sessions, agent memories, workflow history and artifacts.

VALENTINA business records are a separate domain. When a real source workflow demonstrates the need for durable structured storage, one small VALENTINA-owned component may manage one local SQLite database. Do not treat Huginn's internal Pebble store, chat transcripts, Skills or workflow logs as the authoritative business-record database.

This document defines logical JSON contracts for extraction and evaluation. It does not authorize creating a database service before the source workflow justifies it.

## 2. Core entities

### Company

A public-facing organization or publisher associated with one or more opportunities.

Minimum fields: stable company ID, published name, deterministic normalized name, optional website/domain, optional industry tags and public location, source evidence IDs, first-seen and last-seen timestamps.

Do not infer company identity from a similar trading name alone. Preserve alternative names and conflicting evidence rather than silently overwriting.

### Opportunity

One distinct commercial opportunity, not merely a company or web page.

Minimum fields: stable opportunity ID, optional company ID, factual title and summary, owner-configurable opportunity type, review status, optional geography and publication/deadline/validity dates, source evidence IDs, fit assessment tied to a particular owner-profile version, first-seen and last-seen timestamps.

Do not force unknown dates, values, eligibility or location into guessed defaults.

### SourceEvidence

A record that explains where a factual claim came from.

Minimum fields: stable evidence ID, approved source ID, canonical original URL, publisher/page title, observation time, publication time only when present, content hash when available, extractor ID/version, and individual facts with evidence excerpts or locators.

Every material extracted claim must point to at least one evidence record. Distinguish what the publisher states from deterministic derivation and AI interpretation.

## 3. Contact data excluded from v0.1

The v0.1 contract intentionally has no person/contact entity.

Do not collect names, personal email addresses, personal phone numbers, private social profiles or other personal data merely because a page makes them visible. A generic organization-level public contact channel may be considered later only if it is necessary and its purpose, lawful basis, access, retention and deletion rules have been reviewed.

Do not put real contact records or owner-profile values in this public repository.

## 4. Evidence and copyright minimization

Store the canonical source URL, factual fields, timestamps, publisher identity, content hash and only the minimum excerpt needed to explain a claim.

Do not mirror full pages, complete tender dossiers, paywalled documents or large copyrighted texts into the database by default. For PDFs/documents, store the source reference and extracted facts/short evidence locators; retain an original file only when access terms and a concrete operational need allow it.

A search snippet is a discovery hint, not final evidence. Verify against the publisher's original page or official document whenever available.

## 5. Deduplication and entity resolution

Apply matching in this order:

1. Exact source identifier or official opportunity/tender reference.
2. Canonical URL after deterministic tracking-parameter cleanup.
3. Exact normalized title + publisher + publication/deadline window.
4. Strong deterministic company/domain and key-field matches.
5. AI-suggested fuzzy match, which remains a proposal for human review.

Do not merge two opportunities merely because the publisher is the same. One organization can publish many distinct opportunities. Keep duplicate groups and the evidence explaining each match. Destructive merges require human approval and an audit trail.

## 6. Lifecycle

Opportunity states:
- candidate: discovered but not source-verified.
- verified: original source inspected and required facts supported.
- needs_review: ambiguous, conflicting or incomplete.
- qualified: owner-approved as relevant.
- rejected: owner or deterministic hard rule marks it irrelevant/ineligible.
- expired: deadline/validity passed or source confirms closure.
- archived: retained for history but removed from the active queue.

A model may recommend a state; it must not perform an irreversible transition or silently discard a record.

Keep owner decisions, profile version, decision time and reason. Re-assessment after a profile change creates a new assessment rather than overwriting historical reasoning.

## 7. JSON Schema contracts

The provisional machine-readable contracts in schemas/ are:
- company.schema.json
- opportunity.schema.json
- source-evidence.schema.json
- source-profile.schema.json
- owner-profile.schema.json (validates the local-only business profile)
- workflow-input.schema.json (validates the manual research workflow inputs)

Synthetic examples for Company, Opportunity and SourceEvidence are in testdata/fixtures/synthetic/. They contain no real source evidence or business data.

They are data contracts for prompt outputs, fixtures and future validation, not a commitment to the physical SQLite schema. A database design must additionally specify keys, indexes, foreign keys, transactions, migrations, backup and recovery.

## 8. Data boundaries

- Public Git: schemas, placeholder templates, synthetic test fixtures, source-adapter code and documentation only.
- Local private storage: actual owner profile, active source registry where sensitive, operational SQLite database, real opportunities, evidence cache, logs and backups.
- Secrets: Windows Credential Manager/keychain or protected local secret mechanism; never plaintext schema fields, YAML templates, prompts, Git or logs.
- Huginn state: use its native per-user config/data location unless a reviewed migration requires otherwise.

See SECURITY-AND-SECRETS.md and DEPLOYMENT-AND-INSTALLER.md.
