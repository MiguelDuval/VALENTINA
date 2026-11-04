---
name: valentina-extraction-contract
author: VALENTINA
description: Produce evidence-backed opportunity data that conforms to VALENTINA's provisional JSON contracts
huginn:
  priority: 90
---

Extract only information supported by the source text supplied in the task. Use the current schemas in the VALENTINA repository as the contract. Output strict JSON when requested by a workflow; do not wrap it in Markdown fences, commentary or invented defaults.

## Rules

- Follow schemas/opportunity.schema.json, schemas/company.schema.json and schemas/source-evidence.schema.json.
- Never invent an ID, date, price, currency, deadline, publisher, location, eligibility condition or contact detail.
- Use null or omit optional values when unknown, following the schema. Explain the reason in missing-information fields where available.
- Preserve original titles and relevant source language. Summaries may be normalized but must not add claims.
- Distinguish one opportunity from the publisher/company that announced it. One company may have many different opportunities.
- Include a source evidence record and canonical source URL for each candidate.
- Set `demand_evidence_type` for every new candidate: `direct_request`, `public_tender`, `subcontracting_request`, `early_demand_signal` or `unknown`. Use `unknown` when the original evidence does not support a classification.
- Do not confuse the kind of service requested with the kind of demand evidence. An early signal is not an explicit request, tender or awarded job; briefly explain the public evidence in `demand_evidence_note`.
- Keep excerpts short and factual. Do not reproduce a full page, tender document or copyrighted description.
- Use only values from the approved source and owner profile. Do not add unapproved fields.
- Do not mark an item verified unless the original publisher source was successfully inspected.
- If output cannot be validated against the schema, state the validation problem instead of pretending the result is ready for storage.
