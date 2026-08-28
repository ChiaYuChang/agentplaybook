# AgentPlaybook

[![Go Reference](https://pkg.go.dev/badge/github.com/ChiaYuChang/agentplaybook.svg)](https://pkg.go.dev/github.com/ChiaYuChang/agentplaybook)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An on-demand, read-only collaboration manual and protocol reference for multi-agent coding workflows.

Instead of stuffing massive, static role prompts into every agent turn—wasting context tokens and risking cognitive drift—**AgentPlaybook** models workflow knowledge into five orthogonal domains and provides a token-efficient CLI for on-demand progressive disclosure.

---

## Key Features

- **Read-Only Collaboration Manual**: Zero side-effects on your target repository. No orchestrator runtime daemon, no external state store.
- **5 Orthogonal Knowledge Domains**:
  - `role`: Durable participant identities, boundaries, and communication targets (`planner`, `builder`, `reviewer`, `scout`).
  - `flow`: Deterministic multi-agent procedures with semantic condition transitions (`init`, `plan`, `build`, `review`, `commit`).
  - `artifact`: Document and message contracts specifying required sections and visibility boundaries (`agents-md`, `build-plan`, `review-plan`, `review-findings`, `scout-survey`).
  - `rule`: Concrete operational policies and invariants (`anti-cheating`, `mandatory-alignment`, `atomic-change-units`, `tdd-reproduction`, `agents-md-single-writer`, `commit-authority-separation`, `interface-stability-contract-testing`).
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

### 3. Inspecting Flows & Step SOPs
The `init` flow has 9 steps and an optional Scout reconnaissance branch. Step 1 routes `SCOUT_RECON_REQUIRED` to Step 2, where Scout returns a `scout-survey`; `DIRECT_SURVEY` routes directly to Step 3, where Planner validates evidence or surveys the repository before drafting `AGENTS.md`. Reviewer and Builder inquiry convergence then proceeds through Steps 4-8, and Planner finalizes the artifact in Step 9.

```bash
# Full flow definition
agentplaybook flow init
agentplaybook flow commit

# Isolated single step query
agentplaybook flow build --step 2
agentplaybook flow commit --step 5
```

### 4. Inspecting Artifact Contracts
```bash
agentplaybook artifact agents-md
agentplaybook artifact build-plan
agentplaybook artifact review-findings
agentplaybook artifact scout-survey
```

### 5. Inspecting Behavioral Rules
```bash
# List all rule IDs and summaries
agentplaybook rule list

# Explain specific rule IDs
agentplaybook rule explain atomic-change-units
agentplaybook rule explain agents-md-single-writer commit-authority-separation
```

### Interface Stability & Contract Testing
Build plans and implementation diffs must preserve stable component interfaces through explicit boundary declarations and meaningful contract tests:

- A build plan must identify all affected boundary symbols, endpoints, schemas, files, or consumer contracts, or explicitly state that no external boundary is affected.
- Interface changes require a plan amendment before implementation, identifying affected consumers and compatibility or migration handling.
- Contract tests must assert observable input/output, side effects, errors, or interoperability at the boundary, not internal implementation details or mere absence of failure.
- A contract test must fail under at least one plausible violating implementation; Reviewer assesses falsifiability through targeted variation where feasible.
- Unexpected cross-boundary dependencies require Planner escalation; Builder must not unilaterally expand scope.
- Contract tests are distinct from TDD reproduction tests: TDD reproduction is mandatory for validated review findings; contract tests are required when boundary behavior is added, changed, or insufficiently protected.

---

## 3-Tier Architectural Delegation

To eliminate mechanism leakage and preserve strict jurisdictional boundaries (strict separation of concerns):

- **Tier 3 Orchestration (`AgentPlaybook`)**: Purely conceptual multi-agent orchestration protocol. Governs agent roles (`Planner`, `Reviewer`, `Builder`), artifact contracts (`AGENTS.md`, `build-plan`, `review-plan`), and flow state machines (`init`, `plan`, `build`, `review`, `commit`). Defines *what evidence and gates must exist before handoffs are accepted*, remaining completely VCS-neutral and mechanism-agnostic (flow actions contain no raw `jj` or `git` commands).
- **Tier 3 Policy Overlay (`agentcommit`)**: Specialized commit policy overlay skill governing candidate stabilization, TOCTOU verification, secret scanning execution, and authorization checks.
- **Tier 2 Mechanism (`skills/jujutsu` & Git)**: Underlying version control mechanism skills. Concrete tool invocations, command flags, headless guards (`--no-pager`), temporary workspace routines, and conflict resolutions reside exclusively in mechanism skills.
- **Tier 1 Tooling**: Raw binaries (`jj`, `git`, `gitleaks`).

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
