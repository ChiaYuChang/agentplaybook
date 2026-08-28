# AGENTS.md

## Architectural Topology & Jurisdictions

- **Repository Tier**: Tier 3 Orchestration Protocol (`AgentPlaybook`). Roles: `planner`, `reviewer`, `builder`, `scout`. Flows: `init`, `plan`, `build`, `review`, `commit`. Memory: living `AGENTS.md`.
- **External Interfaces**: Go CLI (`agentplaybook`) discovery commands (`role`, `flow`, `artifact`, `rule`) with JSON output.
- **Jurisdictional Boundaries (strict separation of concerns)**:
  - `AgentPlaybook`: Conceptual, evidence-based governance. Strictly VCS-neutral; no raw shell scripts or command syntax.
  - VCS Mechanism: Low-level mechanics, headless guards (`--no-pager`), workspace management delegated to active VCS skill (`skills/jujutsu` or Git).
  - Policy Overlay: Commit candidate stabilization, TOCTOU defense, secret scanning delegated to active commit policy overlay (`agentcommit`).

## Global Operational Invariants

- **Non-Interactive Execution**: Headless-safe only. Prohibit interactive TUIs, unshielded pagers, confirmation prompts in unattended sessions.
- **Living Memory Single-Writer**: Planner sole author/curator of `AGENTS.md`. Builder, Reviewer, Scout never edit directly.
- **Language Standard & Telegraphic Style**: Machine-facing memory in concise en-US ASCII. Drop articles/filler/prose. Non-ASCII domain terms require explicit adjacent inline rationale. Exact symbols/paths mandatory.
- **Inter-Agent Messaging**: Efficiency-first. Drop pleasantries, social framing, human prose. Transmit compact, structured technical payloads with exact symbols/paths.
- **Commit & Publication Separation**: Human commit auth = local seal only. Remote push requires separate explicit user auth.
- **Fail-Closed Intent Recovery**: On `AUTHORIZATION_DENIED`, return to Step 2 awaiting renewed user intent. Autonomous re-drafting forbidden.
- **Conventional Commits**: Messages follow Conventional Commits specification (`feat`, `fix`, `refactor`, `test`, `docs`, `chore`).

## Builder Precautions & Gotchas

- **CLI Role Discovery Sole Truth**: Tracked `roles/*.md` removed in v0.2.0. Query roles via `agentplaybook role <name>` backed by `internal/data/roles.json`.
- **Step Sequence Validation**: `validate.go` requires linear steps to sequence to `Index+1`. Conditional branches (e.g. `DIRECT_SURVEY -> 3`, `PUBLICATION_AUTHORIZED -> 9`) require explicit condition targets.
- **Go Embed Data Invalidation**: `internal/data/*.json` embedded via `embed.go`. Syntax errors invalidate full CLI test suite.
- **`rtk git diff` Path Scope**: Include `--no-ext-diff` before `--` to prevent path filters being parsed as revisions.
- **`flow commit` Non-Mutating Command**: `agentplaybook flow commit` queryable workflow metadata only; coordinator flow, not mutating binary CLI command.
- **`AGENTPLAYBOOK_DEV=1` Cache Race**: Concurrent CLI queries with `AGENTPLAYBOOK_DEV=1` trigger build race on cache (`text file busy`). Run development queries sequentially.

## Reviewer Precautions & Checklist

- **Single Source of Truth**: `internal/data/*.json` canonical CLI truth. Sync flow, artifact, role, rule, docs contracts; cover via matrix tests.
- **VCS-Neutral Language**: Flow step actions and descriptions must remain conceptual/evidence-based; zero embedded `jj`/`git` commands.
- **Public-Only Guidance**: `AGENTS.md` contains public operational guidance only. Exclude private review criteria, hidden test fixtures, inspection techniques.
- **Contract Test Falsifiability**: Boundary contract tests assert observable behavior and must fail on plausible violating implementation; distinct from TDD reproductions.
- **Scout Survey Evidence & Confidentiality**: Verify `scout-survey` provenance, evidence paths, uncertainty markers. Keep Scout read-only; never grant access to private review artifacts.
- **Language Purity & Telegraphic Audit**: During Step 5 commit checks, audit `AGENTS.md` for secret leaks, unauthorized non-ASCII text lacking inline rationale, and conversational fluff.

## Active State & In-Flight Context

- **Observed-At**: `2026-08-28T14:12:00Z @ 5c4329605aa444ee28ebb987af76e958703efec4`
- **Dirty Status**: Modified in-scope files for `v0.2.4` (`internal/data/roles.json`, `internal/data/rules.json`, `internal/cli/matrix_test.go`, `scripts/VERSION`, `README.md`, `SKILL.md`, and `AGENTS.md`).
- **Milestone**: `v0.2.4` - Telegraphic Token-Dense AGENTS.md, Inter-Agent Caveman Protocol & Acknowledgements Governance.
- **Next Pickup Item**: Conclude Governed Commit Flow Step 5 (Reviewer narrow visibility gate), Step 6 (snapshot secret scanning), Step 7 (human commit authorization), and Step 8 (local revision sealing).
- **Ground Truth Revalidation Invariant**: Cold-start Planners MUST run fresh `git status` or `jj --no-pager status` to revalidate mutable repository ground truth; never blindly trust cached Active State.
