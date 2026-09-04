package cli

// DefaultLivingMemoryTemplate returns the canonical AGENTS.md template markdown
// incorporating the 6 roles, 8 flows, dual formal gates, and peer-session primacy over subagents.
func DefaultLivingMemoryTemplate() string {
	return `# AGENTS.md

## Architectural Topology & Jurisdictions

- **Repository Tier**: Tier 3 Orchestration Protocol (` + "`" + `AgentPlaybook v0.3.4 Living Memory Blueprint` + "`" + `). Roles: ` + "`" + `planner` + "`" + `, ` + "`" + `reviewer` + "`" + `, ` + "`" + `builder` + "`" + `, ` + "`" + `scout` + "`" + ` (category: ` + "`" + `core` + "`" + `), ` + "`" + `navigator` + "`" + `, ` + "`" + `cartographer` + "`" + ` (category: ` + "`" + `companion` + "`" + `). Flows: ` + "`" + `init` + "`" + `, ` + "`" + `plan` + "`" + `, ` + "`" + `blueprint` + "`" + `, ` + "`" + `build` + "`" + `, ` + "`" + `review` + "`" + `, ` + "`" + `commit` + "`" + `, ` + "`" + `cartography` + "`" + `, ` + "`" + `session-handoff` + "`" + `. Memory: living ` + "`" + `AGENTS.md` + "`" + `.
- **External Interfaces**: Go CLI (` + "`" + `agentplaybook` + "`" + `) discovery commands (` + "`" + `role` + "`" + `, ` + "`" + `flow` + "`" + `, ` + "`" + `artifact` + "`" + `, ` + "`" + `rule` + "`" + `) and scaffolding (` + "`" + `init` + "`" + `) with JSON/markdown output.
- **Artifact Governance**: Hierarchical structure with ` + "`" + `blueprint-plan` + "`" + ` (` + "`" + `<slug>.blueprint.md` + "`" + `), ` + "`" + `sub-build-plan` + "`" + ` (` + "`" + `sub/<slug>.build.md` + "`" + `), ` + "`" + `sub-review-plan` + "`" + ` (` + "`" + `sub/<slug>.review.md` + "`" + `), ` + "`" + `sub-review-resolution` + "`" + ` (` + "`" + `sub/<slug>.resolution.md` + "`" + `), top-level ` + "`" + `review-resolution` + "`" + ` (` + "`" + `<slug>.resolution.md` + "`" + `), ` + "`" + `diagram-brief` + "`" + `, and ` + "`" + `diagram-completion` + "`" + `.
- **Blind Barrier, Scout Isolation & Companion Allowlist**: ` + "`" + `review-findings` + "`" + ` strictly restricted to ` + "`" + `["planner", "reviewer"]` + "`" + `; Builder receives only Planner-sanitized remediation instructions. Scout strictly excluded from all task in-flight artifacts (` + "`" + `build-plan` + "`" + `, ` + "`" + `review-plan` + "`" + `, ` + "`" + `blueprint-plan` + "`" + `, ` + "`" + `sub-*` + "`" + `, ` + "`" + `review-findings` + "`" + `). Navigator and Cartographer visibility strictly constrained by Settled-Artifact Allowlist (` + "`" + `agents-md` + "`" + `, ` + "`" + `review-resolution` + "`" + `, ` + "`" + `sub-review-resolution` + "`" + `); in-flight draft plans and review artifacts strictly exclude companions. Navigator is never an artifact owner or flow actor; Cartographer owns only ` + "`" + `diagram-completion` + "`" + ` and acts only in ` + "`" + `cartography` + "`" + ` flow.
- **Navigator Companion Governance**:
  - Star-Topology Isolation: Communicates strictly with ` + "`" + `user` + "`" + `, ` + "`" + `planner` + "`" + `, and ` + "`" + `cartographer` + "`" + `. Direct communication with ` + "`" + `builder` + "`" + `, ` + "`" + `reviewer` + "`" + `, ` + "`" + `scout` + "`" + ` strictly forbidden.
  - Zero Instruction Relay: Returns fixed handoff (*"Please send this requirement directly to Planner"*) on change requests.
  - Planner Zero Side-Effect & Zero Response Obligation: Queries never trigger autonomous tasks, plan creation, or repository mutations. Planner possesses complete permission to ignore companion queries.
  - Planner Source-Restricted Response: Responses constrained to facts independently derivable from public allowlist with mandatory ` + "`" + `[Source: <path> | Observed: <rev> @ <timestamp>]` + "`" + ` provenance citation; denylist non-inference enforced.
  - Target-State Gated Inquiry: Queries gated strictly to eligible states (` + "`" + `idle` + "`" + ` or ` + "`" + `done` + "`" + `); non-eligible states prohibit dispatch; recipient discard on arrival, no retry/queue, admission limits (max 1 in-flight, <500 chars payload, fallback to static artifacts).
- **Cartographer Companion Governance**:
  - Specialized Visual Architect: Transforms architectural semantics, system topology, and execution flows into self-contained HTML/inline SVG diagrams under the editorial design system (` + "`" + `docs/diagrams/<safe-name>.html` + "`" + `).
  - Zero Context Pollution: Isolates raw markup within Cartographer session; returns strictly the lightweight ` + "`" + `diagram-completion` + "`" + ` message artifact (<100 tokens evaluated by deterministic subword estimator EstimateTokenCount, <=250 chars, <=60 words, single-sentence digest, zero inline markup) containing persistent file URI, single-sentence plain text summary digest, and node/edge statistics.
  - Taste Gate & Advisory Pushback: Enforces visual suitability and strict complexity budgets (<=12 nodes, <=12 transitions); issues advisory pushback (` + "`" + `ADVISORY_ISSUED` + "`" + `) recommending tables/prose when superior.
  - Asynchronous Fire-and-Forget Decoupling: Commissioners dispatch ` + "`" + `diagram-brief` + "`" + ` asynchronously without blocking or active polling; pipeline flows never gate on cartography operations.
  - Prerequisite Notice: Prompts operator upon startup/role assumption that the ` + "`" + `diagram-design` + "`" + ` skill (https://github.com/cathrynlavery/diagram-design) is required.
  - Strict Star-Topology Isolation: Communicates strictly with ` + "`" + `user` + "`" + `, ` + "`" + `planner` + "`" + `, and ` + "`" + `navigator` + "`" + `. Direct communication with ` + "`" + `builder` + "`" + `, ` + "`" + `reviewer` + "`" + `, ` + "`" + `scout` + "`" + ` strictly forbidden.
- **Jurisdictional Boundaries (strict separation of concerns)**:
  - ` + "`" + `AgentPlaybook` + "`" + `: Conceptual, evidence-based governance. Strictly VCS-neutral; no raw shell scripts or command syntax in catalog data.
  - VCS Mechanism: Low-level mechanics, headless guards (` + "`" + `--no-pager` + "`" + `), workspace management delegated to active VCS skill (Jujutsu / ` + "`" + `agentjj` + "`" + ` or Git).
  - Policy Overlay: Commit candidate stabilization, TOCTOU defense, secret scanning delegated to active commit policy overlay (` + "`" + `agentcommit` + "`" + `).

## Global Operational Invariants

- **Peer-Session Primacy over Subagents**: Reviewer, Builder, Scout, and Cartographer operate as dedicated peer sessions in external panes or workspaces (orchestrated via the active harness transport, e.g. herdr). Planners MUST NEVER spawn nested subagents (e.g. invoke_subagent) to simulate Reviewer or Builder gates. All review dispatches and build tasks MUST be routed to dedicated peer panes to preserve the Blind Barrier and prevent context window exhaustion.
- **Dual Formal Gates**: Two distinct validation gates govern the pipeline: the Plan-Review Gate (` + "`" + `PLAN_PASS` + "`" + `) before any implementation begins, and the Code-Review Gate (` + "`" + `REVIEW_PASS` + "`" + `) before commit synthesis. Neither gate may be bypassed.
- **Non-Interactive Execution**: Headless-safe only. Prohibit interactive TUIs, unshielded pagers, confirmation prompts in unattended sessions.
- **Living Memory Single-Writer**: Planner sole author/curator of ` + "`" + `AGENTS.md` + "`" + `. Builder, Reviewer, Scout, Navigator, Cartographer never edit directly.
- **Language Standard & Telegraphic Style**: Machine-facing memory in concise en-US ASCII. Drop articles/filler/prose. Non-ASCII domain terms require explicit adjacent inline rationale. Exact symbols/paths mandatory.
- **Inter-Agent Messaging**: Efficiency-first. Drop pleasantries, social framing, human prose. Transmit compact, structured technical payloads with exact symbols/paths.
- **Commit & Publication Separation**: Human commit auth = local seal only. Remote push requires separate explicit user auth.
- **Fail-Closed Intent Recovery**: On ` + "`" + `AUTHORIZATION_DENIED` + "`" + `, return to Step 2 awaiting renewed user intent. Autonomous re-drafting forbidden.
- **Conventional Commits**: Messages follow Conventional Commits specification (` + "`" + `feat` + "`" + `, ` + "`" + `fix` + "`" + `, ` + "`" + `refactor` + "`" + `, ` + "`" + `test` + "`" + `, ` + "`" + `docs` + "`" + `, ` + "`" + `chore` + "`" + `). Concise header without plan slug or ` + "`" + `00_` + "`" + ` prefix.
- **No Background Auto-Update**: Updates user-initiated only. Runner never polls, downloads, or mutates repository in background without explicit invocation.
- **Coherent Plan Units & Anti-Rubber-Stamp Gate**: Sub-plans authored Just-In-Time (JIT) from high-level Blueprint plans. Single change request decomposed into coherent sub-plans by architectural layers, bounded contexts, or independent invariants. Reviewer must independently challenge excessive consolidation.
- **Finding Severity**: Exactly ` + "`" + `Blocker` + "`" + `, ` + "`" + `Major` + "`" + `, ` + "`" + `Minor` + "`" + `, ` + "`" + `Other` + "`" + `. Unresolved Blocker blocks ` + "`" + `REVIEW_PASS` + "`" + `; Major requires resolution or documented Planner waiver; Minor/Other non-blocking.
- **Verification Tracks**: Track A covers local behavioral RED/GREEN. Non-behavioral findings use static/specification evidence. Optional Track B covers complete action differentials with pinned baseline identity tuple ` + "`" + `(repository_identity, baseline_identity)` + "`" + ` and fails closed (` + "`" + `BASELINE_STALE` + "`" + `) on baseline drift.
- **Zero Context Pollution**: Cartographer contains raw HTML/SVG markup within isolated session; returns strictly lightweight ` + "`" + `diagram-completion` + "`" + ` (<100 tokens, <=250 chars, <=60 words, single-sentence plain text summary digest, zero inline markup).

## Builder Precautions & Gotchas

- **CLI Role Discovery Sole Truth**: Query roles via ` + "`" + `agentplaybook role <name>` + "`" + ` backed by ` + "`" + `internal/data/roles.json` + "`" + `.
- **Step Sequence Validation**: ` + "`" + `validate.go` + "`" + ` requires linear steps to sequence to ` + "`" + `Index+1` + "`" + `. Conditional branches require explicit condition targets.
- **Go Embed Data Invalidation**: ` + "`" + `internal/data/*.json` + "`" + ` embedded via ` + "`" + `embed.go` + "`" + `. Syntax errors invalidate full CLI test suite.
- **` + "`" + `rtk git diff` + "`" + ` Path Scope**: Include ` + "`" + `--no-ext-diff` + "`" + ` before ` + "`" + `--` + "`" + ` to prevent path filters being parsed as revisions.
- **` + "`" + `flow commit` + "`" + ` Non-Mutating Command**: ` + "`" + `agentplaybook flow commit` + "`" + ` queryable workflow metadata only; coordinator flow, not mutating binary CLI command.
- **` + "`" + `AGENTPLAYBOOK_DEV=1` + "`" + ` Cache Race**: Concurrent CLI queries with ` + "`" + `AGENTPLAYBOOK_DEV=1` + "`" + ` trigger build race on cache (` + "`" + `text file busy` + "`" + `). Run development queries sequentially.
- **Stateless Replaceability**: Builder stateless/disposable. Bloated or rate-limited sessions replaced from approved build plans without compaction token overhead.
- **Artifact Data Shape**: Message artifacts use ` + "`" + `ArtifactField` + "`" + `; document artifacts use ` + "`" + `ArtifactSection` + "`" + `. JSON syntax or missing embedded rule/artifact references invalidate the CLI test suite.
- **Non-Behavioral Review Evidence**: Docs, schema, and contract-rule findings use static/specification evidence and must not create artificial RED test harnesses.

## Reviewer Precautions & Checklist

- **Single Source of Truth**: ` + "`" + `internal/data/*.json` + "`" + ` canonical CLI truth. Sync flow, artifact, role, rule, docs contracts; cover via matrix tests.
- **VCS-Neutral Language**: Flow step actions and descriptions must remain conceptual/evidence-based; zero embedded ` + "`" + `jj` + "`" + `/` + "`" + `git` + "`" + ` commands in catalog data.
- **Public-Only Guidance**: ` + "`" + `AGENTS.md` + "`" + ` contains public operational guidance only. Exclude private review criteria, hidden test fixtures, inspection techniques.
- **Contract Test Falsifiability**: Boundary contract tests assert observable behavior and must fail on plausible violating implementation; distinct from TDD reproductions.
- **Plan Coverage**: Audit every material Build Plan invariant against an independent Review Plan verification path; inspect optional Track B fields, baseline identity, and severity disposition semantics.
- **Resolution Hygiene**: Shared ` + "`" + `review-resolution` + "`" + ` and ` + "`" + `sub-review-resolution` + "`" + ` are Planner-sanitized and actionable-only; exclude review-plan criteria, hidden fixtures, and private inspection methods. Keep task-specific findings out of ` + "`" + `AGENTS.md` + "`" + `.
- **Scout Survey Evidence & Confidentiality**: Verify ` + "`" + `scout-survey` + "`" + ` provenance, evidence paths, uncertainty markers. Keep Scout read-only; never grant access to private review artifacts or in-flight task plans.
- **Language Purity & Telegraphic Audit**: During Step 5 commit checks, audit ` + "`" + `AGENTS.md` + "`" + ` for secret leaks, unauthorized non-ASCII text lacking inline rationale, and conversational fluff.
- **Post-Commit Compaction Self-Evaluation**: After completing commit flow, self-evaluate context window; execute compaction when usage > 50% or review clutter accumulates on high-tier model.

## Active State & In-Flight Context

- **Observed-At**: ` + "`" + `YYYY-MM-DDTHH:MM:SSZ @ revision` + "`" + `
- **Dirty Status**: ` + "`" + `clean / dirty (<description>)` + "`" + `
- **Milestone**: ` + "`" + `Milestone Title - Status (e.g. READY_TO_INIT)` + "`" + `
- **Next Pickup Item**: ` + "`" + `Immediate next action item` + "`" + `
- **Ground Truth Revalidation Invariant**: Cold-start Planners MUST run fresh VCS status inspection to revalidate mutable repository ground truth; never blindly trust cached Active State.
`
}

// MinimalLivingMemoryTemplate returns the ultra-compact telegraphic Caveman-style AGENTS.md template markdown.
// Strictly adheres to size budget (<=50 lines, <=2500 bytes, <25% of DefaultLivingMemoryTemplate) and pure ASCII.
func MinimalLivingMemoryTemplate() string {
	return `# AGENTS.md

## Topology
- Tier: 3. Core: planner, reviewer, builder, scout. Companion: navigator, cartographer. Memory: AGENTS.md.
- Flow: init, plan, blueprint, build, review, commit, cartography, session-handoff.
- Artifacts: blueprint-plan, sub-build-plan, sub-review-plan, sub-review-resolution, review-resolution, diagram-brief, diagram-completion.
- Companion Allowlist: Companions see only agents-md, review-resolution, sub-review-resolution. No draft plans or review-findings access.
- Star Topology: Builder, Reviewer, Scout talk ONLY to Planner. Navigator/Cartographer talk ONLY to User, Planner, Navigator/Cartographer.
- Navigator: Handoff: "Please send this requirement directly to Planner". Provenance: [Source: <path> | Observed: <rev> @ <timestamp>]. Gated: idle/done only; discard on arrival, no retry/queue, max 1 in-flight, <500 chars, static fallback. Zero side-effects.
- Cartographer: Write: docs/diagrams/<safe-name>.html (traversal prohibited). Return: diagram-completion (file URI, single-sentence digest, node/edge statistics; <100 tokens, <=60 words, <=250 runes, zero inline markup). Taste Gate: budget <=12 nodes, <=12 transitions; pushback ADVISORY_ISSUED. async non-blocking. Prerequisite: diagram-design.

## Invariants
- Peer-Session Primacy: Reviewer, Builder, Scout, Cartographer = external panes via herdr. NEVER call invoke_subagent for Reviewer/Builder. Protects Blind Barrier and context.
- Dual Gates: PLAN_PASS before code, REVIEW_PASS before commit. No bypass.
- Headless: Non-interactive only. No TUIs/pagers.
- Single Writer: Planner curates AGENTS.md. Others never edit.
- Commit: Human auth = local seal only. Remote push requires separate auth.
- Fail-Closed: Return to Step 2 on AUTHORIZATION_DENIED.
- Format: Pure ASCII, telegraphic, drop filler words.

## Builder Rules
- Single truth: Query CLI agentplaybook role/flow/rule.
- Embed data: JSON syntax error invalidates CLI suite.
- Blind barrier: Never inspect review-plan.
- No VCS: Code and test only. No git/jj commits.

## Reviewer Checklist
- Independent verification: Run tests, race, vet, gofmt, diff check.
- Severity: Blocker (blocks pass), Major (resolve or waiver), Minor/Other (non-blocking).
- Plan vs Code: Challenge scope creep and missing invariants.

## Active State
- Observed-At: <TIMESTAMP> @ <REV>
- Status: Clean
- Milestone: Active work
- Next: Awaiting user intent
`
}
