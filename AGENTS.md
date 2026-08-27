# AGENTS.md

## Architectural Topology & Jurisdictions

- **Repository Tier**: Tier 3 Orchestration Protocol (`AgentPlaybook`) governing multi-agent collaboration, durable roles (`planner`, `reviewer`, `builder`), deterministic flows (`init`, `plan`, `build`, `review`, `commit`), and memory lifecycle contracts.
- **External Interfaces**: Go CLI binary (`agentplaybook`) providing discovery-friendly commands (`role`, `flow`, `artifact`, `rule`) with JSON output.
- **Jurisdictional Boundaries ("不宜跨區辦案")**:
  - `AgentPlaybook` maintains purely conceptual, evidence-based governance, remaining strictly VCS-neutral without embedding raw shell scripts or command syntax.
  - Low-level VCS mechanics, headless guards (`--no-pager`), and workspace management are delegated to the active VCS mechanism skill (`/home/cychang/Projects/agent/skills/jujutsu` or Git).
  - Commit candidate stabilization, TOCTOU defense, and secret scanning are governed via the active commit policy overlay (`agentcommit`).

## Global Operational Invariants

- **Non-Interactive Execution**: All operations must be headless-safe; interactive TUIs, unshielded pagers, and interactive confirmation prompts are prohibited in unattended sessions.
- **Living Memory Single-Writer Principle**: Planner is the sole author and curator of `AGENTS.md`. Builder and Reviewer must never edit `AGENTS.md` directly.
- **Commit & Publication Authority Separation**: Commit authorization permits local revision sealing only. Remote publication requires separate, explicit human authorization.
- **Fail-Closed Intent Recovery**: Upon `AUTHORIZATION_DENIED`, Planner must return to Step 2 to await renewed user intent; autonomous re-drafting is strictly forbidden.
- **Conventional Commits**: Commit messages follow Conventional Commits specification (`feat`, `fix`, `refactor`, `test`, `docs`, `chore`).

## Builder Precautions & Gotchas

- **CLI Role Discovery as Sole Truth**: Tracked `roles/*.md` files have been deprecated and removed in `v0.2.0`. All role queries must be conducted via `agentplaybook role <name>` backed by `internal/data/roles.json`.
- **Sequential Step Validation**: In `internal/data/flows.json`, `validate.go` requires steps without conditions to sequence to `Index+1`. Intermediate steps with conditional branching (such as Step 8 `PUBLICATION_AUTHORIZED -> 9`) must declare explicit condition targets to satisfy validation.
- **Go Embed Data Invalidation**: `internal/data/*.json` files are embedded via Go `embed.go`. All edits to embedded JSON must be strictly well-formed, as JSON syntax errors invalidate test execution across the entire CLI suite.
- **`rtk git diff` Path Scope**: When running path-scoped `git diff` via `rtk`, include `--no-ext-diff` before `--`; otherwise paths may be misinterpreted as revision arguments rather than path filters.
- **`flow commit` Not a Mutating CLI Command**: The 9-step governed commit flow is queryable via `agentplaybook flow commit` but its execution is strictly Planner-governed coordination. Do not attempt to invoke it as a standalone mutating CLI command; execute each step under Planner authority.

## Reviewer Precautions & Checklist

- **Single Source of Truth**: Treat `internal/data/*.json` as the CLI's source of truth. Keep flow, artifact, role, rule, and documentation contracts synchronized and covered by matrix tests.
- **VCS-Neutral Language**: Verify that all flow step actions and descriptions remain purely conceptual and evidence-based without embedding raw `jj` or `git` command strings.
- **Public-Only Guidance**: Ensure `AGENTS.md` contains only public, independently observable operational guidance, strictly excluding confidential review criteria, hidden test fixtures, or private inspection techniques.
- **Contract Test Falsifiability**: Boundary contract tests must verify observable behavior and be falsifiable under a plausible violating implementation; they do not replace TDD reproductions for review findings.

## Active State & In-Flight Context

- **Observed-At**: `2026-08-27T21:35:00Z @ fb5fa49b1bdf30f4610f55130b481ee238b4ea3d`
- **Dirty Status**: Modified in-scope files for `v0.2.1` (`internal/data/roles.json`, `internal/data/rules.json`, `internal/cli/matrix_test.go`, `scripts/VERSION`, `README.md`, `SKILL.md`, and `AGENTS.md`).
- **Milestone**: `v0.2.1` - Component Interface Stability & Meaningful Contract Testing Governance.
- **Next Pickup Item**: Conclude Governed Commit Flow Step 5 (Reviewer narrow visibility gate), Step 6 (snapshot secret scanning), Step 7 (human commit authorization), and Step 8 (local revision sealing).
- **Ground Truth Revalidation Invariant**: Receiving Planners cold-starting a session MUST execute fresh VCS inspection commands (`git status` or `jj --no-pager status`) to revalidate mutable repository ground truth rather than blindly trusting this Active State block.
