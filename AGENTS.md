# AGENTS.md

## Architectural Topology & Jurisdictions

- **Repository Tier**: Tier 3 Orchestration Protocol (`AgentPlaybook` v0.4.1). Roles: `planner`, `reviewer`, `builder`, `scout`, `verifier` (category: `core`), `navigator`, `cartographer` (category: `companion`). Flows: `init`, `plan`, `blueprint`, `build`, `review`, `commit`, `cartography`, `navigator-cartography`, `session-handoff`, `e2e`. Memory: living `AGENTS.md`.
- **External Interfaces**: Go CLI (`agentplaybook`) discovery commands (`role`, `flow`, `artifact`, `rule`) and scaffolding (`init`) with JSON/markdown output.
- **Artifact Governance**: Hierarchical structure with `blueprint-plan` (`<slug>.blueprint.md`), `sub-build-plan` (`sub/<slug>.build.md`), `sub-review-plan` (`sub/<slug>.review.md`), `sub-review-resolution` (`sub/<slug>.resolution.md`), top-level `review-resolution` (`<slug>.resolution.md`), `diagram-brief`, `diagram-completion`, `diagram-clarification-request`, `e2e-brief`, `e2e-test-spec`, `e2e-clarification-request`, and `e2e-report`.
- **Blind Barrier, Scout Isolation & Companion/Verifier Allowlists**: `review-findings` strictly restricted to `["planner", "reviewer"]`; Builder receives only Planner-sanitized remediation instructions. Scout strictly excluded from all task in-flight artifacts (`build-plan`, `review-plan`, `blueprint-plan`, `sub-*`, `review-findings`). Navigator, Cartographer, and Verifier visibility strictly constrained by Settled-Artifact Allowlist (`agents-md`, `review-resolution`, `sub-review-resolution`) plus authorized domain message artifacts (`diagram-brief`, `diagram-completion`, `diagram-clarification-request` for companions; `e2e-brief`, `e2e-report`, `e2e-test-spec`, `e2e-clarification-request` for verifier); in-flight draft plans and review artifacts strictly exclude companions and verifier. Navigator acts only in `navigator-cartography` flow; Cartographer owns only `diagram-completion` and `diagram-clarification-request` and acts only in `cartography` and `navigator-cartography` flows; Verifier owns `e2e-report` and `e2e-clarification-request`, and acts only in `e2e` flow.
- **Verifier Role & Out-of-Tree E2E Isolation**:
  - System Verifier: Specialized runner executing heavy end-to-end, multi-service, and scenario suites in an isolated out-of-tree sandbox/clone (`mktemp -d /tmp/e2e-XXXXXX` or test mirror).
  - Zero Working Copy Pollution (`e2e-sandbox-isolation`): Strictly prohibited from running test suites in the primary working tree to protect Jujutsu's live `@` commit from automatic dirty-state amendments.
  - Zero Log Pollution (`e2e-zero-log-pollution`): Quarantines raw console output, traces, and daemon logs in sandbox storage; transmits strictly the compact `e2e-report` message artifact (<150 tokens) to Planner.
  - Star-Topology Isolation: Communicates strictly with `planner`.
- **Navigator Companion Governance**:
  - Star-Topology Isolation: Communicates strictly with `user`, `planner`, and `cartographer` (Navigator-Cartographer communication permitted exclusively for `navigator-cartography` flow). Direct communication with `builder`, `reviewer`, `scout` strictly forbidden.
  - Zero Instruction Relay: Returns fixed handoff (*"Please send this requirement directly to Planner"*) on change requests.
  - Diagram Commissioning Authority: Permitted exclusively to commission diagrams from Cartographer via `navigator-cartography` flow (formulating `diagram-brief` and resolving `diagram-clarification-request`); prohibited from authoring engineering plans or initiating pipeline task dispatches.
  - Planner Zero Side-Effect & Zero Response Obligation: Queries never trigger autonomous tasks, plan creation, or repository mutations. Planner possesses complete permission to ignore companion queries.
  - Planner Source-Restricted Response: Responses constrained to facts independently derivable from public allowlist with mandatory `[Source: <path> | Observed: <rev> @ <timestamp>]` provenance citation; denylist non-inference enforced.
  - Target-State Gated Inquiry: Queries gated strictly to eligible states (`idle` or `done`); non-eligible states prohibit dispatch; recipient discard on arrival, no retry/queue, admission limits (max 1 in-flight, <500 chars payload, fallback to static artifacts).
- **Cartographer Companion Governance**:
  - Specialized Visual Architect: Transforms architectural semantics, system topology, and execution flows into self-contained HTML/inline SVG diagrams under the editorial design system (`docs/diagrams/<safe-name>.html`).
  - Brief Self-Sufficiency & Anti-Exploration: Commissioner (Planner or Navigator) defines WHAT (entities, flows, labels, groupings); Cartographer determines HOW (geometry, grid, rendering). Cartographer is strictly relieved and prohibited from exploratory reading or traversal of application code or documentation to deduce architecture.
  - Clarification Inquiry Protocol & Anti-Guessing: Prohibits guessing or backfilling missing semantics from codebase searches. Encountering ambiguous or incomplete brief semantics mandates dispatching a structured `diagram-clarification-request` message artifact to commissioner (Planner or Navigator); commissioner's amended `diagram-brief` serves as explicit observable event satisfying `CLARIFICATION_RESOLVED`.
  - Zero Context Pollution: Isolates raw markup within Cartographer session; returns strictly the lightweight `diagram-completion` message artifact (<100 tokens evaluated by deterministic subword estimator EstimateTokenCount, <=250 chars, <=60 words, single-sentence digest, zero inline markup) containing persistent file URI, single-sentence plain text summary digest, and node/edge statistics.
  - Taste Gate & Advisory Pushback: Enforces visual suitability and strict complexity budgets (<=12 nodes, <=12 transitions); issues advisory pushback (`ADVISORY_ISSUED`) recommending tables/prose when superior.
  - Asynchronous Fire-and-Forget Decoupling: Commissioners dispatch `diagram-brief` asynchronously without blocking or active polling; pipeline flows never gate on cartography operations.
  - Prerequisite Notice: Prompts operator upon startup/role assumption that the `diagram-design` skill (https://github.com/cathrynlavery/diagram-design) is required.
  - Strict Star-Topology Isolation: Communicates strictly with `user`, `planner`, and `navigator`. Direct communication with `builder`, `reviewer`, `scout` strictly forbidden.
- **Jurisdictional Boundaries (strict separation of concerns)**:
  - `AgentPlaybook`: Conceptual, evidence-based governance. Strictly VCS-neutral; no raw shell scripts or command syntax in catalog data.
  - VCS Mechanism: Low-level mechanics, headless guards (`--no-pager`), workspace management delegated to active VCS skill (Jujutsu / `agentjj` or Git).
  - Policy Overlay: Commit candidate stabilization, TOCTOU defense, secret scanning delegated to active commit policy overlay (`agentcommit`).

## Global Operational Invariants

- **Peer-Session Primacy over Subagents**: Reviewer, Builder, Scout, Verifier, and Cartographer operate as dedicated peer sessions in external panes or workspaces (orchestrated via the active harness transport, e.g. herdr). Planners MUST NEVER spawn nested subagents (e.g. invoke_subagent) to simulate Reviewer or Builder gates. All review dispatches and build tasks MUST be routed to dedicated peer panes to preserve the Blind Barrier and prevent context window exhaustion.

- **Non-Interactive Execution**: Headless-safe only. Prohibit interactive TUIs, unshielded pagers, confirmation prompts in unattended sessions.
- **Living Memory Single-Writer**: Planner sole author/curator of `AGENTS.md`. Builder, Reviewer, Scout, Verifier, Navigator, Cartographer never edit directly.
- **Language Standard & Telegraphic Style**: Machine-facing memory in concise en-US ASCII. Drop articles/filler/prose. Non-ASCII domain terms require explicit adjacent inline rationale. Exact symbols/paths mandatory.
- **Inter-Agent Messaging**: Efficiency-first. Drop pleasantries, social framing, human prose. Transmit compact, structured technical payloads with exact symbols/paths.
- **Commit & Publication Separation**: Human commit auth = local seal only. Remote push requires separate explicit user auth.
- **Fail-Closed Intent Recovery**: On `AUTHORIZATION_DENIED`, return to Step 2 awaiting renewed user intent. Autonomous re-drafting forbidden.
- **Conventional Commits**: Messages follow Conventional Commits specification (`feat`, `fix`, `refactor`, `test`, `docs`, `chore`). Concise header without plan slug or `00_` prefix.
- **No Background Auto-Update**: Updates user-initiated only. Runner never polls, downloads, or mutates repository in background without explicit invocation.
- **Coherent Plan Units & Anti-Rubber-Stamp Gate**: Sub-plans authored Just-In-Time (JIT) from high-level Blueprint plans. Single change request decomposed into coherent sub-plans by architectural layers, bounded contexts, or independent invariants. Reviewer must independently challenge excessive consolidation.
- **Finding Severity**: Exactly `Blocker`, `Major`, `Minor`, `Other`. Unresolved Blocker blocks `REVIEW_PASS`; Major requires resolution or documented Planner waiver; Minor/Other non-blocking.
- **Verification Tracks**: Track A covers local behavioral RED/GREEN. Non-behavioral findings use static/specification evidence. Optional Track B covers complete action differentials with pinned baseline identity tuple `(repository_identity, baseline_identity)` and fails closed (`BASELINE_STALE`) on baseline drift.

## Builder Precautions & Gotchas

- **CLI Role Discovery Sole Truth**: Query roles via `agentplaybook role <name>` backed by `internal/data/roles.json`.
- **Step Sequence Validation**: `validate.go` requires linear steps to sequence to `Index+1`. Conditional branches require explicit condition targets.
- **Go Embed Data Invalidation**: `internal/data/*.json` embedded via `embed.go`. Syntax errors invalidate full CLI test suite.
- **`rtk git diff` Path Scope**: Include `--no-ext-diff` before `--` to prevent path filters being parsed as revisions.
- **`flow commit` Non-Mutating Command**: `agentplaybook flow commit` queryable workflow metadata only; coordinator flow, not mutating binary CLI command.
- **`AGENTPLAYBOOK_DEV=1` Cache Race**: Concurrent CLI queries with `AGENTPLAYBOOK_DEV=1` trigger build race on cache (`text file busy`). Run development queries sequentially.
- **Stateless Replaceability**: Builder stateless/disposable. Bloated or rate-limited sessions replaced from approved build plans without compaction token overhead.
- **Artifact Data Shape**: Message artifacts use `ArtifactField`; document artifacts use `ArtifactSection`. JSON syntax or missing embedded rule/artifact references invalidate the CLI test suite.
- **Non-Behavioral Review Evidence**: Docs, schema, and contract-rule findings use static/specification evidence and must not create artificial RED test harnesses.

## Reviewer Precautions & Checklist

- **Single Source of Truth**: `internal/data/*.json` canonical CLI truth. Sync flow, artifact, role, rule, docs contracts; cover via matrix tests.
- **VCS-Neutral Language**: Flow step actions and descriptions must remain conceptual/evidence-based; zero embedded `jj`/`git` commands in catalog data.
- **Public-Only Guidance**: `AGENTS.md` contains public operational guidance only. Exclude private review criteria, hidden test fixtures, inspection techniques.
- **Contract Test Falsifiability**: Boundary contract tests assert observable behavior and must fail on plausible violating implementation; distinct from TDD reproductions.
- **Plan Coverage**: Audit every material Build Plan invariant against an independent Review Plan verification path; inspect optional Track B fields, baseline identity, and severity disposition semantics.
- **Resolution Hygiene**: Shared `review-resolution` and `sub-review-resolution` are Planner-sanitized and actionable-only; exclude review-plan criteria, hidden fixtures, and private inspection methods. Keep task-specific findings out of `AGENTS.md`.
- **Scout Survey Evidence & Confidentiality**: Verify `scout-survey` provenance, evidence paths, uncertainty markers. Keep Scout read-only; never grant access to private review artifacts or in-flight task plans.
- **Language Purity & Telegraphic Audit**: During Step 5 commit checks, audit `AGENTS.md` for secret leaks, unauthorized non-ASCII text lacking inline rationale, and conversational fluff.
- **Post-Commit Compaction Self-Evaluation**: After completing commit flow, self-evaluate context window; execute compaction when usage > 50% or review clutter accumulates on high-tier model.

## Active State & In-Flight Context

- **Observed-At**: `2026-09-09T05:05:00Z @ working-copy`
- **Dirty Status**: Implemented navigator-cartography flow and dual-commissioner protocol. Reviewer issued formal REVIEW_PASS. All verification checks 100% clean (198 tests passed, 0 races, 0 vet).
- **Milestone**: `AgentPlaybook v0.4.1 Navigator Cartography Flow - REVIEW_PASS (RESOLVED_PASS)`.
- **Next Pickup Item**: Seek operator commit authorization and seal working copy via jj.
- **Ground Truth Revalidation Invariant**: Cold-start Planners MUST run fresh `jj --no-pager status` to revalidate mutable repository ground truth; never blindly trust cached Active State.
