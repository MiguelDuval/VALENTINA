---
name: valentina-fit-assessment
author: VALENTINA
description: Evaluate opportunity relevance against explicit business criteria without inventing a score
huginn:
  priority: 80
---

Assess only source-verified facts against the owner profile supplied for this run. The owner's criteria, not generic market assumptions, define relevance. A recommendation is a review aid, not a final business decision.

## Rules

- Evaluate every applicable profile criterion separately as matches, does_not_match or unknown.
- A missing fact is unknown, not a match. Do not infer capacity, eligibility, budget, legal suitability or profitability without evidence.
- Apply a hard disqualifier only when the owner profile explicitly marks it as a hard gate and evidence supports the disqualification.
- Do not create a numeric score or threshold unless the profile explicitly defines its weighting and threshold.
- Explain why a possible match is relevant in plain language.
- Distinguish evidence confidence from commercial attractiveness.
- Treat direct requests, public tenders and subcontracting requests as explicit demand; treat `early_demand_signal` only as indirect evidence of possible future demand. Do not give an indirect signal the same certainty label as an explicit request.
- Apply location priorities from the local owner profile as soft ranking factors unless that profile explicitly marks an area as a hard gate. Do not invent public-transport travel times or assume a private vehicle is available.
- Report contradictory facts and the exact evidence that conflicts.
- Recommend strong_match, possible_match, insufficient_information or not_a_match; do not change record state or discard it.
- Never recommend contacting a named individual unless the owner has explicitly approved that workflow.
