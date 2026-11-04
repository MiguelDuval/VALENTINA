# VALENTINA UI — Design System

Direction: **Deep Slate / Sea Glass**. Target: Windows 11 desktop browser, personal local application.

## Intent
A calm, polished operations workbench: task-first hierarchy, quiet navigation, truthful information and progressive access to advanced capabilities. The first screen should not look like a developer console.

## References
- Linear's March 2026 refresh: consistent headers, easier scanning and a quieter sidebar so the working area takes priority: https://linear.app/changelog/2026-03-12-ui-refresh and https://linear.app/now/behind-the-latest-design-refresh
- Atlassian Design System: semantic interaction tokens, dark-theme mapping and contrast-aware color: https://atlassian.design/foundations/color/
- W3C WCAG 2.2: keyboard focus and contrast: https://www.w3.org/TR/WCAG22/

## Palette
| Token | Value | Use |
|---|---|---|
| valentina-bg | #0A1016 | Canvas |
| valentina-sidebar | #0E171F | Navigation |
| valentina-surface | #141F29 | Cards |
| valentina-panel | #192632 | Raised surfaces |
| valentina-raised | #202F3C | Hover / selected |
| valentina-border | #2A3A47 | Dividers |
| valentina-text | #E8F0F5 | Primary text |
| valentina-muted | #9AAAB7 | Metadata |
| valentina-accent | #49D6C2 | Brand / primary action |
| valentina-blue | #82AAFF | Information / models |
| valentina-violet | #B39CFF | Agents |
| valentina-amber | #F0BD68 | Review / deadlines |
| valentina-green | #59D6A6 | Verified / healthy |
| valentina-red | #FF7E88 | Failed / blocked |

Legacy huginn-* Tailwind classes remain aliases to avoid mass rewrites. New first-party components use valentina-* tokens. Status uses text/icons as well as color.

## Typography and layout
- Inter if installed; otherwise Segoe UI Variable / Segoe UI / system UI. No network font downloads.
- Monospace only for code and IDs.
- 224 px sidebar on desktop; 72 px compact icon rail at narrower widths.
- 24–32 px content padding; 8 px spacing rhythm; 12–16 px card radius.
- Visible keyboard focus and accessible names for icon-only controls.

## Navigation
Обзор, Ассистент, Агенты, Память, Модели, Процессы, Подключения, Навыки, Статистика, Журнал, Уведомления, Настройки. Keep existing engine features available. Future Возможности, Компании and Источники sections must not appear as working features until real routes and data contracts exist.

## Overview and future CRM
Overview values come from existing endpoints: health, agents, enabled workflows and pending notifications. Until opportunity storage and a real source connector exist, show an honest empty state rather than invented leads. Future opportunities show company, work categories, demand evidence type, municipality, travel priority, deadline/freshness, fit and human review state. Original source and evidence remain visible; travel times must never be guessed.

## Boundary
This patch changes the theme, shell, title, favicon, navigation and Overview. It does not implement CRM persistence, source discovery, verified NIM tool/streaming compatibility, full localization, Playwright visual QA or NSIS installer.

## Upstream safety
First-party changes are stored as an ordered patch outside the pinned submodule. Windows build copies the exact source SHA into a temp directory, verifies/applies patches, runs frontend tests and builds the embedded Go executable. It never edits the submodule or uploads the local-use executable.
