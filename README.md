# AgentPlaybook

[![Go Reference](https://pkg.go.dev/badge/github.com/ChiaYuChang/agentplaybook.svg)](https://pkg.go.dev/github.com/ChiaYuChang/agentplaybook)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An on-demand, read-only collaboration manual and protocol reference for multi-agent coding workflows.

Instead of stuffing massive, static role prompts into every agent turn—wasting context tokens and risking cognitive drift—**AgentPlaybook** models workflow knowledge into five orthogonal domains and provides a token-efficient CLI for on-demand progressive disclosure.

---

## Key Features

- **Read-Only Collaboration Manual & Opt-In Scaffolding**: AgentPlaybook remains an evidence-based, read-only guidance manual; the 'agentplaybook init' command is an explicit, opt-in local scaffolding utility executed strictly upon operator invocation to generate baseline AGENTS.md, with zero background mutation, network downloads, or daemon processes.
- **5 Orthogonal Knowledge Domains**:
  - `role`: Core and companion participant identities, boundaries, and communication targets (`planner`, `builder`, `reviewer`, `scout`, `navigator`, `cartographer`).
  - `flow`: Deterministic multi-agent procedures with semantic condition transitions (`init`, `plan`, `blueprint`, `build`, `review`, `commit`, `session-handoff`, `cartography`).
  - `artifact`: Document and message contracts specifying required sections and visibility boundaries (`agents-md`, `build-plan`, `review-plan`, `blueprint-plan`, `sub-build-plan`, `sub-review-plan`, `sub-review-resolution`, `review-findings`, `scout-survey`, `review-resolution`, `diagram-brief`, `diagram-completion`).
  - `rule`: Concrete operational policies and invariants (`anti-cheating`, `mandatory-alignment`, `coherent-plan-units`, `anti-rubber-stamp-plan-gate`, `evidence-proportional-persistence`, `tdd-reproduction`, `agents-md-single-writer`, `acceptance-publication-authority`, `interface-stability-contract-testing`, `session-handoff-audit`, `planner-reviewability`, `review-severity-semantics`, `track-b-action-differential-verification`, `out-of-tree-baseline-mirror`, `navigator-read-only-companion`, `companion-query-zero-side-effect`, `planner-source-restricted-response`, `target-state-gated-inquiry`, `cartographer-visual-architect-boundary`, `cartography-zero-context-pollution`, `cartography-taste-gate-advisory`, `cartography-asynchronous-decoupling`, `peer-session-transport-primacy`).
  - `config`: Supported languages, prefix templates, and transport settings.
- **Progressive Disclosure UX**: Bare discovery commands output concise catalogs (Exit 0); specific queries return clean, indented JSON.
- **Built for AI Agents**: Automatic self-caching runner script, compatible with [skills.sh](https://skills.sh) across 17+ agent harnesses.

---

## Installation

### Method 1: AI Agent Harnesses (via [skills.sh](https://skills.sh))

Install directly into any supported agent harness (Claude Code, Cursor, Antigravity, OpenCode, Codex, Cline, Amp, etc.):

```bash
# Install to current project
npx skills add ChiaYuChang/agentplaybook

# Or install globally across all agents
npx skills add ChiaYuChang/agentplaybook -g
```

### Method 2: Global Go CLI

```bash
go install github.com/ChiaYuChang/agentplaybook@latest
```

### Method 3: Local Clone

```bash
git clone https://github.com/ChiaYuChang/agentplaybook.git
cd agentplaybook
go build -o bin/agentplaybook .
```

### Prebuilt Release Binaries

`scripts/run-agentplaybook.sh` downloads the matching Linux or macOS (`amd64` or `arm64`) release archive when no versioned cache exists. Before extraction, it requires exactly one matching archive record in `checksums.txt` and verifies the archive with SHA-256. Download, platform, or checksum failures fall back to a local `go build` when Go is available.

Use `--update` or `AGENTPLAYBOOK_UPDATE=1` to force-refresh the prebuilt binary from GitHub Releases, bypassing the cache while retaining checksum verification and atomic replacement. If the refresh fails, the runner falls back to a local `go build`. Use `--build` or `AGENTPLAYBOOK_BUILD=1` to bypass a cached prebuilt binary and atomically replace it with a fresh local build:

```bash
scripts/run-agentplaybook.sh --update --version
AGENTPLAYBOOK_UPDATE=1 scripts/run-agentplaybook.sh --help
scripts/run-agentplaybook.sh --build --version
AGENTPLAYBOOK_BUILD=1 scripts/run-agentplaybook.sh --help
```

Set `AGENTPLAYBOOK_RELEASE_BASE_URL` to override the release download base URL for mirrors or controlled fixtures. `AGENTPLAYBOOK_DEV=1` continues to build directly from the local workspace.

Runner progress is written to `stderr`; CLI output on `stdout` remains clean for piping. Warm cache hits are silent. Cache misses report download, SHA-256 verification, and successful caching. `--update` downloads only `checksums.txt` first and compares its archive hash with the cached `.archive_hash`; matching releases report `already up-to-date` without downloading the archive, while changed releases are downloaded, verified, and atomically replaced. Local builds remove `.archive_hash` before compilation.

### Installation and Transport

`skills.sh` provides universal skill installation and loading across 17+ agent harnesses, including Antigravity, Claude Code, Cursor, OpenCode, Codex, and others. Multi-agent transport is separate: it governs real-time communication between agent sessions and is currently expected to be provided by Herdr. Installing a skill via skills.sh does not provide the transport layer.

### VCS Recommendation

Jujutsu is the preferred VCS for multi-agent collaboration because native change stacking supports isolated working-copy revisions and clear handoffs. Git is fully supported; use the same Planner-owned version control governance and verified diff handoff with either backend.

---

## CLI Usage

### 1. Discovery (Bare Commands)
Bare invocations print concise text catalogs and exit 0:
```bash
agentplaybook           # Overall manual
agentplaybook role      # List participant roles
agentplaybook flow      # List available orchestration flows
agentplaybook artifact  # List document contracts and schemas
agentplaybook rule      # List rule subcommands
```

### 2. Inspecting Roles & Boundaries

> **Breaking Change (v0.2.0)**: The legacy `roles/` directory (`roles/*.md`) has been removed. Role definitions, boundaries, responsibilities, and communication targets are inspected exclusively via the `agentplaybook role <name>` CLI command backed by embedded knowledge datasets.

```bash
# Full role definition (JSON)
agentplaybook role builder

# Query specific sections
agentplaybook role builder --responsibility
agentplaybook role builder --boundary
agentplaybook role builder --communication
```

### Scout Role & Read-Only Reconnaissance
The optional `scout` role performs read-only reconnaissance of large or unfamiliar repositories. Scout maps directory structure, entry points, build graphs, component boundaries, symbols, dependencies, and toolchains, then delivers factual findings to Planner without modifying repository files or persistent documentation.

```bash
agentplaybook role scout
agentplaybook role scout --responsibility
agentplaybook role scout --boundary
agentplaybook role scout --communication
```

Scout findings use the transient `scout-survey` message artifact. It records the survey pass ID, provenance, repository topology, module boundaries, build and toolchain observations, concrete evidence, and uncertainties. Planner validates the evidence against live repository ground truth before synthesizing it into `AGENTS.md` or task plans; Scout never edits `AGENTS.md` and must not access reviewer-only artifacts.

Model capacity routing is advisory rather than mandatory: the default recommendation is `Scout >= Reviewer >= Planner >= Builder`, with allocation scaled by task domain, scope, uncertainty, and risk.

### Navigator Role & Comprehension Companion
The `navigator` companion role (`category: "companion"`) accompanies the human operator in co-reading and exploring the codebase, generating visual call graphs (Mermaid), and translating complex multi-file diffs into human-friendly change digests while maintaining read-only isolation, admission controls, and transport silence toward engineering pipelines.

```bash
agentplaybook role navigator
agentplaybook role navigator --responsibility
agentplaybook role navigator --boundary
agentplaybook role navigator --communication
```

Navigator operates under a strict star topology communicating only with the user, Planner, and Cartographer. On detecting user change intent, Navigator executes a fixed handoff: *"Please send this requirement directly to Planner"*.

### 3. Inspecting Flows & Step SOPs
The `init` flow has 9 steps and an optional Scout reconnaissance branch. Step 1 begins with active workspace peer-session discovery through the active harness transport, prioritizing dedicated Reviewer, Builder, or Scout peer sessions rather than spawning nested subagents, resolves active Tier 2 VCS capability, then routes `SCOUT_RECON_REQUIRED` to Step 2, where Scout returns a `scout-survey`; `DIRECT_SURVEY` routes directly to Step 3, where Planner validates evidence or surveys the repository before drafting `AGENTS.md`. Reviewer and Builder inquiry convergence then proceeds through Steps 4-8, and Planner finalizes the artifact in Step 9.

The `blueprint` flow defines a 12-step deterministic hierarchical feature lifecycle: Planner authors `<slug>.blueprint.md` (Step 1), Reviewer executes the Blueprint Gate (Step 2), Planner authors JIT sub-plan pairs `sub/<slug>.build.md` and `sub/<slug>.review.md` (Step 3), Reviewer executes the Sub-Plan Gate (Step 4), Builder implements the sub-build plan (Step 5), Reviewer inspects diffs and reports findings (Step 6), Planner mediates and sanitizes findings (Step 7), Planner synthesizes `sub/<slug>.resolution.md` (Step 8), Reviewer verifies sub-resolution (Step 9), Planner synthesizes master composition `<slug>.resolution.md` (Step 10), Reviewer executes the Feature Composition Gate (Step 11), and Planner triggers Governed Milestone Acceptance (Step 12).

The `session-handoff` flow defines an 8-step protocol for session transition, anchor capture, and takeover readiness verification. Step 1 performs active workspace peer-session discovery through the active harness transport and prioritizes dedicated Reviewer, Builder, or Scout peer sessions rather than spawning nested subagents, then captures the State & Topology Anchor. Step 2 revalidates anchor state against live repository ground truth. Steps 3-5 execute 3 structured readiness audit rounds across all four mandatory dimensions (Plan Understanding, Progress & State Understanding, Architectural & Governance Decisions, and Permissions & Path-Scoping Invariants). Step 6 enforces tool root path-scoping and blind barrier invariants, Step 7 validates Plan-Review Gate independence and release cadence, and Step 8 emits the validated takeover state. The manual provides read-only guidance with no runtime transport mutation.

```bash
# Full flow definition
agentplaybook flow init
agentplaybook flow plan
agentplaybook flow blueprint
agentplaybook flow session-handoff
agentplaybook flow commit

# Isolated single step query
agentplaybook flow blueprint --step 2
agentplaybook flow build --step 2
agentplaybook flow session-handoff --step 1
agentplaybook flow commit --step 5
```

### 4. Inspecting Artifact Contracts
```bash
agentplaybook artifact agents-md
agentplaybook artifact build-plan
agentplaybook artifact blueprint-plan
agentplaybook artifact sub-build-plan
agentplaybook artifact sub-review-plan
agentplaybook artifact sub-review-resolution
agentplaybook artifact review-findings
agentplaybook artifact scout-survey
agentplaybook artifact review-resolution
```

### 5. Inspecting Behavioral Rules
```bash
# List all rule IDs and summaries
agentplaybook rule list

# Explain specific rule IDs
agentplaybook rule explain coherent-plan-units anti-rubber-stamp-plan-gate
agentplaybook rule explain evidence-proportional-persistence acceptance-publication-authority
agentplaybook rule explain agents-md-single-writer session-handoff-audit
```

### Hierarchical Blueprint & Governance Architecture
AgentPlaybook v0.3.0 establishes hierarchical blueprinting and two-tier gating for complex feature engineering:

1. **Coherent Plan Units (`coherent-plan-units`)**:
   One plan owns exactly one root intent. Internal steps decompose execution sequence rather than acceptance atomicity. Decomposition uses domain-driven signals: coupling, cross-cutting scope, mixed concerns, and verification heterogeneity. Multi-stage features branch into hierarchical blueprint plans with JIT sub-plans.
2. **Anti-Rubber-Stamp Plan Gate (`anti-rubber-stamp-plan-gate`)**:
   Reviewer actively executes the Counterfactual Decomposition Challenge on proposed plans (`SPLIT_ATTEMPT` + `SPLIT_REJECTED_BECAUSE`). Plans exhibiting unreviewable coupling or missing verification paths receive short-circuit structural rejection.
3. **Evidence-Proportional Persistence (`evidence-proportional-persistence`)**:
   `plan.md` and `review.md` are mandatory execution contracts; sanitized resolution documents (`sub-review-resolution` and `review-resolution`) capture durable outcomes and verified evidence at gate convergence; `AGENTS.md` is updated strictly when durable cross-task invariants, toolchain gotchas, or project rules emerge.
4. **Milestone Acceptance & Finalization Equivalence (`acceptance-publication-authority`)**:
   Distinguishes `WORKING != ACCEPTED != PUBLISHED`. Milestone acceptance seals reviewed revisions locally as `ACCEPTED` under Planner VCS governance, strictly upholding Finalization Equivalence ($\text{tree}(\text{Final}) == \text{tree}(\text{Verified})$) with zero unstaged drift. Remote publication (`PUBLISHED`) requires explicit separate human authorization.
5. **Two-Tier Gating & Composition Review**:
   - `Blueprint Gate`: Reviewer validates architectural coherence, public contracts, and sub-plan boundary definitions before implementation.
   - `Sub-Plan Gate`: Reviewer validates JIT sub-build and sub-review plan pairs for verification coverage.
   - `Feature Composition Gate`: Reviewer evaluates global composition across all completed sub-plans, shared contracts, and regression suites before milestone acceptance.

### Session Handoff & Takeover Audit Protocol
The `session-handoff-audit` rule formalizes session transitions across 3 core handover elements:
1. **State & Topology Anchor**: Captures `target_roots`, `tier_topology`, `role_assignments`, `baseline_revision`, `dirty_status`, `recent_milestones`, `utc_plan_paths`, `next_pickup`, and `evidence`. Anchor state serves as orienting context and must be revalidated against live repository ground truth.
2. **>= 3 Readiness-Audit Rounds**: Every round verifies Plan Understanding, Progress & State Understanding, Architectural & Governance Decisions, and Permissions & Path-Scoping Invariants.
3. **Path-Scoping Enforcement**: Tool roots are strictly scoped to authorized directories before takeover completion or task dispatch, maintaining the Builder blind barrier.

On Planner startup or workflow initiation, Planner performs active workspace peer-session discovery through the active harness transport and prioritizes dispatch to existing dedicated Reviewer, Builder, or Scout peer sessions over spawning nested subagents. Release cadence is strictly defined: daily bugfix/patch commits roll on `main` without SemVer tag bumps; SemVer tags and release publication are reserved for consolidated milestones and remain explicitly authorized. Flow definitions provide read-only reference guidance with no runtime transport mutation.

### Interface Stability & Contract Testing
Build plans and implementation diffs must preserve stable component interfaces through explicit boundary declarations and meaningful contract tests:

- A build plan must identify all affected boundary symbols, endpoints, schemas, files, or consumer contracts, or explicitly state that no external boundary is affected.
- Interface changes require a plan amendment before implementation, identifying affected consumers and compatibility or migration handling.
- Contract tests must assert observable input/output, side effects, errors, or interoperability at the boundary, not internal implementation details or mere absence of failure.
- A contract test must fail under at least one plausible violating implementation; Reviewer assesses falsifiability through targeted variation where feasible.
- Unexpected cross-boundary dependencies require Planner escalation; Builder must not unilaterally expand scope.
- Contract tests are distinct from TDD reproduction tests: TDD reproduction is mandatory for validated review findings; contract tests are required when boundary behavior is added, changed, or insufficiently protected.

### AI Reviewer Spec & 3-Tier Verification Governance

AgentPlaybook formalizes six review and verification pillars within its conceptual Tier 3 orchestration manual:

1. **Planner Reviewability & Verification Coverage (`planner-reviewability`)**:
   Planner owns the reviewability of implementation units. Work decomposition uses domain signals (structural coupling, cross-cutting scope, mixed concerns, verification heterogeneity) rather than an arbitrary or rigid line-count threshold like `>400 LOC`. Every material invariant, boundary condition, and acceptance criterion in the Build Plan must receive a corresponding, independently verifiable path in the Review Plan.
2. **Formal Severity Classification (`review-severity-semantics`)**:
   Review findings are classified strictly into four formal severity levels:
   - `Blocker`: Severe defect or boundary violation. An unresolved Blocker strictly prohibits `REVIEW_PASS`, requiring implementation resolution or explicit Planner arbitration.
   - `Major`: Significant defect or contract gap. Requires implementation resolution before advancing, unless Planner explicitly records a documented waiver with rationale.
   - `Minor`: Advisory suggestion or non-blocking improvement.
   - `Other`: General observation, architectural discussion, or alternative approach.
   The ambiguous term `Critical` is strictly excluded from severity classification.
3. **Track A Local Behavioral Verification (`tdd-reproduction`)**:
   Governs local behavioral defect fixes through single-function, local-logic test reproduction (`RED -> GREEN`). Non-behavioral review findings (such as documentation, linting, static policies, naming, or design comments) use verified static or specification evidence without requiring artificial failing test harnesses.
4. **Track B Action-Differential Verification (`track-b-action-differential-verification`)**:
   Optional verification track spanning complete action boundaries between Unit and E2E for two distinct purposes: performance-quality characterization and supply-chain/security boundary audit. Planning defines the action, boundary, `baseline_identity`, environment, metrics, sampling, and criteria. Measurements employ interleaved distribution sampling evaluating median, p95, and variance. Any unexplained persistent differential between candidate and baseline identity requires formal investigation before review approval.
5. **Out-of-Tree Baseline Mirror Governance (`out-of-tree-baseline-mirror`)**:
   Planner maintains an out-of-tree, persistent, reconstructible baseline mirror representing pinned baseline identity B0. Reviewer performs read-only baseline identity assertions. If candidate base revision drifts from B0, verification fails closed with `BASELINE_STALE`. Mirror advancement occurs strictly upon local revision sealing (`ACCEPTED`-only, post-seal), maintaining the invariant that `ACCEPTED != PUBLISHED`.
6. **Structured Review Artifacts & Resolution (`review-resolution`)**:
   - `review-plan`: extended with an optional `Track B Definition` section.
   - `review-findings`: requires `evidence_mode` and concrete `evidence`, making `reproduction_scenario` conditional for behavioral findings only.
   - `review-resolution`: Planner-owned document persisted at `plan/{timestamp}/{slug}.resolution.md`, synthesized and sanitized by Planner before shared visibility across Planner, Builder, Reviewer, and Navigator. Contains five required sections: `Outcome`, `Resolved Findings`, `Deviations & Rationales`, `Residual Risks`, and `Verification Evidence`. Shared resolution artifacts strictly exclude confidential review criteria, hidden test fixtures, and private inspection techniques.

### Navigator Companion Role & Zero Side-Effect Governance

AgentPlaybook v0.3.1 introduces the `navigator` companion role and four zero side-effect governance rules:

1. **Navigator Read-Only Companion & Fixed Handoff (`navigator-read-only-companion`)**:
   Navigator operates as a read-only comprehension companion to the human user. On detecting user intent for code changes or workflow dispatch, Navigator executes a fixed handoff: *"Please send this requirement directly to Planner"*, rather than relaying instructions or drafting plans.
2. **Companion Query Zero Side-Effect (`companion-query-zero-side-effect`)**:
   Inquiries from Navigator to Planner are informational only and strictly non-actionable. Companion inquiries never trigger tasks, plans, subagents, or repository mutations. Planner possesses zero response obligation (complete permission to ignore or drop queries without failing any workflow gate).
3. **Planner Source-Restricted Response & Provenance (`planner-source-restricted-response`)**:
   When Planner elects to answer a companion query, responses are strictly restricted to facts independently derivable from Navigator's public allowlist (source code, `AGENTS.md`, `sub-review-resolution`, `review-resolution`). Quoting or inferring from confidential review plans or findings is prohibited. Every factual response requires mandatory provenance citation: `[Source: <path> | Observed: <rev> @ <timestamp>]`.
4. **Target-State Gated Inquiry & Admission Control (`target-state-gated-inquiry`)**:
   Companion inquiries to Planner are gated by recipient lifecycle status: dispatch is permitted only when Planner is in eligible states (`idle` or `done`). Non-eligible states (`running`, `busy`, `waiting_for_input`) strictly prohibit dispatch; inquiries arriving during non-eligible states are discarded immediately without queueing or retry. Navigator enforces admission control: maximum 1 in-flight inquiry, payload under 500 characters, and fallback to static artifacts.

### Operator Guidelines for Navigator & Companion Roles
- **Comprehension & Exploration**: Ask Navigator for call graphs, architecture explanations, or diff digests at any time.
- **Feature Requests & Code Modifications**: Send requirements directly to Planner. If sent to Navigator, expect the fixed handoff: *"Please send this requirement directly to Planner"*.
- **Pipeline Silence**: Navigator queries are lightweight and zero side-effect, respecting Planner's active workflow state (`idle`/`done`) and admission controls.

### Cartographer Companion Role & Visual Architecture Governance

The `cartographer` companion role (`category: "companion"`) is a specialized visual architect dedicated to transforming architectural semantics, system topology, and execution flows into publication-grade, self-contained HTML/inline SVG diagrams under the editorial design system, while guaranteeing zero context pollution and non-blocking asynchronous decoupling for engineering pipelines.

> [Notice] Cartographer requires the diagram-design skill (https://github.com/cathrynlavery/diagram-design) for publication-grade diagram rendering.

#### Core Cartography Rules & Protocols
1. **Path-Scoping & Visual Boundaries (`cartographer-visual-architect-boundary`)**:
   Cartographer operates strictly read-only on application source code and test suites. Write permissions are strictly confined to `docs/diagrams/<safe-name>.html`, rejecting directory traversal (`..`), backslashes, leading slashes, and external paths. Cartographer never edits `AGENTS.md` directly and is prohibited from authoring build plans or participating in review gates.
2. **Taste Gate & Complexity Budget (`cartography-taste-gate-advisory`)**:
   Before rendering, Cartographer exercises a **Taste Gate** assessment to evaluate visual suitability against strict complexity budgets (≤12 nodes for structural diagrams, ≤12 transitions for sequence flows). When markdown tables or structured prose provide superior clarity, Cartographer issues an advisory pushback (`ADVISORY_ISSUED`) directly to Step 5 without rendering markup.
3. **Zero Context Pollution Invariant (`cartography-zero-context-pollution`)**:
   Raw HTML and inline SVG diagram markup remain strictly contained within Cartographer's isolated session and are persisted directly to `docs/diagrams/<safe-name>.html`. Cross-session handoff to commissioners (Planner or Navigator) is strictly restricted to the lightweight `diagram-completion` message artifact (<100 tokens evaluated by deterministic subword estimator EstimateTokenCount, <=250 chars, <=60 words, single-sentence digest, zero inline markup), containing persistent file URI, single-sentence plain text summary digest, and node/edge statistics with zero raw markup.
4. **Asynchronous Decoupling (`cartography-asynchronous-decoupling`)**:
   Commissioners dispatch `diagram-brief` artifacts asynchronously in fire-and-forget mode. Synchronous waiting loops or sleep polling on Cartographer completion are strictly prohibited, ensuring active engineering pipelines are never blocked by diagram synthesis.

#### Cartography Flow (`flow cartography`)
1. Planner identifies visualization requirement and formulates `diagram-brief`.
2. Planner dispatches `diagram-brief` message artifact to Cartographer.
3. Cartographer evaluates visual suitability and complexity budget under Taste Gate criteria (`DIAGRAM_APPROVED` -> Step 4, `ADVISORY_ISSUED` -> Step 5).
4. Cartographer calculates layout geometry, renders HTML/SVG, and performs conceptual visual validation.
5. Cartographer persists diagram to `docs/diagrams/<safe-name>.html` (or records advisory proposal) and emits lightweight `diagram-completion` message artifact to commissioner.

---

## 3-Tier Architectural Delegation

To eliminate mechanism leakage and preserve strict jurisdictional boundaries (strict separation of concerns):

- **Tier 3 Orchestration (`AgentPlaybook`)**: Purely conceptual multi-agent orchestration protocol. Governs agent roles (`Planner`, `Reviewer`, `Builder`), artifact contracts (`AGENTS.md`, `build-plan`, `review-plan`), and flow state machines (`init`, `plan`, `build`, `review`, `commit`). Defines *what evidence and gates must exist before handoffs are accepted*, remaining completely VCS-neutral and mechanism-agnostic (flow actions contain no raw `jj` or `git` commands).
- **Tier 3 Policy Overlay ([`agentcommit`](https://github.com/ChiaYuChang/agentcommit))**: Specialized commit policy overlay skill governing candidate stabilization, TOCTOU verification, secret scanning execution, and authorization checks.
- **Tier 2 Mechanism (Jujutsu & Git VCS Mechanism Skills)**: Underlying version control mechanism skills. Concrete tool invocations, command flags, headless guards (`--no-pager`), temporary workspace routines, and conflict resolutions reside exclusively in mechanism skills.
- **Tier 1 Tooling**: Raw binaries (`jj`, `git`, `gitleaks`).

---

## Living Memory Initialization & Scaffolding (`agentplaybook init`)

AgentPlaybook remains an evidence-based, read-only guidance manual; the 'agentplaybook init' command streams the standard AGENTS.md template to stdout by default (zero filesystem writes). When invoked with `--file`, it serves as an explicit, opt-in local scaffolding utility executed strictly upon operator invocation to generate baseline AGENTS.md, with zero background mutation, network downloads, or daemon processes.

```bash
# Stream template directly to stdout (zero disk writes, read-only inspection)
agentplaybook init

# Initialize standard AGENTS.md in current directory (fails if file already exists)
agentplaybook init --file AGENTS.md

# Force overwrite existing AGENTS.md (enforces strict 0644 permissions)
agentplaybook init --file AGENTS.md --force
# or
agentplaybook init -f AGENTS.md -F

# Scaffold to custom destination with automatic directory creation
agentplaybook init --file docs/AGENTS.md
```

### Peer-Session Primacy over Subagents (`peer-session-transport-primacy`)
To prevent recursive context compaction churn and preserve the Blind Barrier, AgentPlaybook enforces **Peer-Session Primacy over Subagents**:
- Reviewer, Builder, Scout, and Cartographer operate as dedicated peer sessions in external panes or workspaces (orchestrated via the active harness transport, e.g. herdr).
- Planners MUST NEVER spawn nested subagents (e.g. `invoke_subagent`) to simulate Reviewer or Builder gates.
- All review dispatches and build tasks MUST be routed to dedicated peer panes to preserve the Blind Barrier and prevent anti-compaction churn.

---

## Living `AGENTS.md` Operational Memory

Instead of fragmented append-only journal files that risk creating competing truths and $O(N)$ cold-start overhead, `AgentPlaybook` adopts a single, living `AGENTS.md` maintained at the repository root as an "Agent-Facing README" for instant $O(1)$ session bootstrapping.

### Canonical 5 Sections
1. **Architectural Topology & Jurisdictions**: Repository tier, external interfaces, and jurisdictional boundaries (strict separation of concerns).
2. **Global Operational Invariants**: Project-level invariants (e.g., non-interactive execution guards, Conventional Commits, branch conventions).
3. **Builder Precautions & Gotchas**: Toolchain quirks, compiler limitations, and test runner constraints.
4. **Reviewer Precautions & Checklist**: Public verification guidelines and regression checkpoints (strictly excluding confidential test secrets).
5. **Active State & In-Flight Context**: Pre-commit baseline observation with mandatory provenance (`Observed-At: <UTC timestamp> @ <base-revision-id>`), dirty status, recent milestones, and next pickup item.

### Single-Writer & Ground Truth Invariants
- **Single-Writer Principle**: Planner is the sole author and curator of `AGENTS.md`. Builder and Reviewer are strictly prohibited from editing it directly.
- **Live VCS Revalidation**: Active State provides orienting context, never a substitute for live repository ground truth. Any receiving Planner cold-starting a session MUST execute fresh VCS inspection commands (`status`/`log`) to revalidate mutable facts before planning or executing tasks.
- **Blind-Barrier Check**: During the commit flow, Reviewer conducts a narrow visibility check on `AGENTS.md` to ensure no confidential review criteria or hidden fixtures leak into shared documentation (`BARRIER_LEAK` returns to Planner for redaction).

### Language & Conciseness Standard
- `AGENTS.md` must be authored in concise US English (`en-US`) and remain pure ASCII by default.
- A non-ASCII domain term is permitted only with an explicit adjacent inline rationale, such as `(domain term: concise rationale)`; this is the inline exception contract.
- Planner keeps entries concise and high-signal, excluding unnecessary verbosity, redundant explanations, and unverified operational narratives.
- Reviewer audits the file during the commit flow for unauthorized non-ASCII text lacking inline justification, excessive verbosity or bloat, and confidential barrier leakage.

### Telegraphic Memory & Inter-Agent Communication
- `AGENTS.md` is machine-facing operational memory and a token-dense cache for LLM agents, not human-facing prose. Use concise fragments; drop articles, filler words, conversational pleasantries, and redundant grammar while preserving exact technical terms and code symbols.
- Inter-agent communication requires telegraphic (caveman) compression for all exchanges (including [Planner], [Reviewer], [Builder], [Scout], or [<role>] prefixed messages), omitting conversational filler and pleasantries in favor of compact structured technical fragments.
- Telegraphic formatting is an endogenous specification only; it introduces no new runtime packages, binaries, or dependencies.

### Role-Tiered Context Lifecycle & Compaction Governance
- Reviewer operates on high-tier reasoning models; after completing a commit flow, self-evaluate context utilization and compact or clean up when usage exceeds 50% or accumulated review history bloats operating costs.
- Builder operates as a stateless, disposable worker; replace bloated or quota-limited sessions with a fresh instance from the approved `<slug>.plan.md` rather than spending tokens on compaction.
- Planner retains orchestration history, curates living memory, and anchors transitions across ephemeral role lifecycles.

## Acknowledgements & Prior Art
- **Caveman Prompting Pattern**: Julius Brussee, [`github.com/JuliusBrussee/caveman`](https://github.com/JuliusBrussee/caveman), for the telegraphic token-dense compression paradigm for LLMs.
- **Jujutsu (`jj`)**: [`github.com/jj-vcs/jj`](https://github.com/jj-vcs/jj), the Git-compatible VCS with first-class change stacking and working-copy snapshots.
- **Gitleaks**: [`github.com/gitleaks/gitleaks`](https://github.com/gitleaks/gitleaks), the automated audit and secret detection engine used for pre-commit verification.
- **Test-Driven Development & Contract Testing**: Kent Beck, *Test-Driven Development: By Example* (Addison-Wesley, 2002); Ian Robinson, *Consumer-Driven Contracts: A Service Evolution Pattern* (2006), [`martinfowler.com/articles/consumerDrivenContracts.html`](https://martinfowler.com/articles/consumerDrivenContracts.html).

These specifications introduce no new runtime dependencies.

---

## 9-Step Governed Commit Flow (`flow commit`)

A conceptual, evidence-based commit pipeline modeling candidate stabilization, living memory updates, security verification, and separate human authorization:

```text
[1. Confirm Review Approval (REVIEW_PASS)]
                    │
[2. Await Explicit User Commit Request] ◄──────────────┐
                    │                                  │ AUTHORIZATION_DENIED
[3. Query Operational Caveats (Builder + Reviewer)]    │
                    │                                  │
[4. Update AGENTS.md (Planner sole writer)] ◄────┐     │
                    │                            │     │
[5. Reviewer Visibility Check] ─── BARRIER_LEAK ─┤     │
                    │ AGENTS_REVIEW_PASS         │     │
[6. Candidate Stabilization & Secret Scan] ── SCAN_FAILED
                    │ SCAN_CLEAN                       │
[7. Present Diff & Conventional Commit] ───────────────┘
                    │ AUTHORIZATION_GRANTED
[8. Seal Revision Locally (Planner VCS Governance)]
                    │ PUBLICATION_AUTHORIZED
[9. Publish Sealed Revision to Remote Repository]
```

- **Commit & Publication Authority Separation**: Human commit authorization permits local revision sealing only; publishing to a remote repository requires separate, explicit publication authorization.
- **Fail-Closed Intent Recovery**: If commit authorization is denied (`AUTHORIZATION_DENIED`), Planner must return to Step 2 to await renewed user intent; autonomous re-drafting is strictly forbidden.

---

## Architecture & Design Principles

```text
                 agentplaybook CLI (Read-Only)
                              │
    ┌─────────────┬───────────┴───────────┬─────────────┐
    ▼             ▼                       ▼             ▼
  role          flow                   artifact       rule
(Identity)   (Procedures)             (Contracts)   (Invariants)
```

- **Separation of Governance and Craft**: The playbook defines *who* validates, *what* sections are required, and *which* boundaries are strictly enforced. Agents use their own reasoning and toolsets to fulfill the contracts.
- **Zero Runtime Interference**: The CLI never spawns agents, makes network transport calls, or mutates repository state.

---

## License

MIT
