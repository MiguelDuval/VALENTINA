# VALENTINA UI — implementation status

## Implemented and verified locally

- VALENTINA dark palette, Russian primary navigation and Overview shell.
- CRM routes for Opportunities, Opportunity Detail, Leads, Companies, Sources and Research.
- Opportunities list uses server-side status filtering and bounded pagination.
- Opportunity cards open a dedicated detail page.
- Opportunity detail loads persisted facts, geography, dates/value, linked SourceEvidence, fit assessment and entity-scoped review history.
- Source URLs use safe external-link attributes (`noopener noreferrer`).
- Operator review actions require a reason and submit through the transactional API; status transitions are constrained by the domain state machine.
- Leads retains manual creation; Companies, Sources and Research retain their existing list views.
- Loading, error, not-found and empty states are present on the opportunity workflow.

## Local verification — 2026-11-18

- Vitest: 115 files / 2,002 tests passed.
- Vue typecheck passed.
- Vite production build passed (832 modules transformed).
- Go domain/API tests passed.
- Upstream server compile check passed.
- Full Windows x64 build passed.
- EXE: 66,955,776 bytes.
- SHA-256: `2E1135B318863F6F419DDC0781DA5142B923DBC662EC0566354593C3512EA447`.
- Isolated runtime smoke: health 200; unauthenticated CRM 401; opportunity list 200; source evidence 200; invalid review filters 400; missing opportunity 404.

The EXE and ZIP were produced in a unique temporary output directory. No existing `dist` artifact or user EXE was overwritten. The real Huginn profile was not used by the final runtime smoke.

## Not complete / not certified

- InboundLead detail and review screen remains a P2 follow-up.
- Companies, Sources and Research still need deeper operator workflows and visual polish.
- No source connector, crawling, scraping, evidence collection or research-run execution is enabled; these are P3 and remain blocked.
- NVIDIA NIM function calling/streaming is not certified; P4 remains blocked.
- Playwright visual regression and full Windows browser interaction tests remain.
- NSIS installer remains a later milestone.
- GitHub Actions has not yet verified this unpushed local checkpoint; the local artifact is not a published release.

## Build

Run from repository root:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-valentina.ps1 -Profile Full -OutputDirectory <unique-temp-directory>
```

The Full profile verifies the upstream pin, overlays first-party CRM packages into a disposable upstream copy, runs frontend and Go checks, compiles the Windows executable and packages the artifact. It must not modify `upstream/huginn`.
