---
name: agentplaybook
description: Use the AgentPlaybook Collaboration Manual CLI to coordinate multi-agent workflows, retrieve role boundaries, procedural flows, artifact contracts, and behavioral rules.
---

# AgentPlaybook Collaboration Manual CLI

The AgentPlaybook CLI is a read-only informational manual and reference playbook for multi-agent collaboration.
Use it as the definitive source of truth for participant roles, end-to-end orchestration flows, document contracts, and operational rules.

Commands:

- Linux or macOS: `sh "<skill-dir>/scripts/run-agentplaybook.sh"`
- Local development: `AGENTPLAYBOOK_DEV=1 sh "<skill-dir>/scripts/run-agentplaybook.sh"`

## Core Philosophy

The AgentPlaybook CLI is an informational collaboration manual. It provides on-demand reference for roles, flows, artifact contracts, and rules.
The manual queries perform no workflow mutations, track no live state, spawn no agents, and execute no transport calls. It has zero side-effects on the target repository.
Interpret the retrieved guidance and perform the work using your own reasoning and tools.

## Progressive Disclosure Protocol

Do not query every knowledge domain on every turn. Query only the specific domain required for your immediate context:

1. **Role & Identity**:
   When establishing your participant identity or checking your allowed boundaries and responsibilities (the CLI is the sole authoritative source of truth; legacy `roles/` markdown files are removed):
   ```sh
   sh "<skill-dir>/scripts/run-agentplaybook.sh" role <role-name>
   # Selectors: --description, --responsibility, --boundary, --communication
   ```

2. **Flow & Procedures**:
   When entering a collaboration phase or determining the next step in a sequence:
   ```sh
   sh "<skill-dir>/scripts/run-agentplaybook.sh" flow <flow-name>
   # Query single step: flow <flow-name> --step <index>
   ```

3. **Artifact Contracts**:
   Before authoring, reviewing, or exchanging persistent plans, summaries, or structured findings:
   ```sh
   sh "<skill-dir>/scripts/run-agentplaybook.sh" artifact <artifact-name>
   ```

4. **Rules & Protocols**:
   When encountering specific boundary situations, test reproduction protocols, or invariant checks:
   ```sh
   sh "<skill-dir>/scripts/run-agentplaybook.sh" rule list
   sh "<skill-dir>/scripts/run-agentplaybook.sh" rule explain <rule-id>...
   ```

## Scout Role & Read-Only Reconnaissance

The optional `scout` role performs strictly read-only reconnaissance for large or unfamiliar repositories. Scout maps directory structure, entry points, build graphs, component boundaries, factual symbols, dependencies, and toolchains, then sends a transient `scout-survey` message to Planner. Scout must not edit, create, or delete repository files, modify VCS state, or edit `AGENTS.md`.

The `scout-survey` artifact contains seven required fields: `id`, `provenance`, `repository_topology`, `module_boundaries`, `build_and_toolchains`, `evidence`, and `uncertainties`. Planner validates those findings against live repository ground truth before synthesizing them into `AGENTS.md` or task plans. Scout must not read, search for, or request task-specific review plans, reviewer-only tests, or verification artifacts.

Model capacity routing is advisory: the default recommendation is `Scout >= Reviewer >= Planner >= Builder`, with model allocation scaled by task domain, scope, uncertainty, and risk.

## Repository Initialization Reconnaissance

The `init` flow has nine steps with an optional Scout branch:

1. Planner assesses repository scale and chooses `SCOUT_RECON_REQUIRED` -> Step 2 or `DIRECT_SURVEY` -> Step 3.
2. Scout performs read-only topological reconnaissance and returns a structured `scout-survey` to Planner.
3. Planner validates Scout evidence against repository ground truth, or performs the direct survey, and drafts `AGENTS.md`.
4. Reviewer reviews the current `AGENTS.md` artifact.
5. Planner incorporates Reviewer feedback.
6. Builder reviews the updated `AGENTS.md` artifact.
7. Planner incorporates Builder feedback.
8. Planner evaluates consensus: `QUESTIONS_RAISED` -> Step 4 or `NO_QUESTIONS_RAISED` -> Step 9.
9. Planner finalizes and persists `AGENTS.md`.

## Interface Stability & Contract Testing

The `interface-stability-contract-testing` rule governs component boundaries and the tests that protect them:

- A build plan must identify all affected boundary symbols, endpoints, schemas, files, or consumer contracts, or explicitly state that no external boundary is affected.
- Interface changes require a plan amendment before implementation, identifying affected consumers and compatibility or migration handling.
- Contract tests must assert observable input/output, side effects, errors, or interoperability at the boundary, not internal implementation details or mere absence of failure.
- A contract test must fail under at least one plausible violating implementation; Reviewer assesses falsifiability through targeted variation where feasible.
- Unexpected cross-boundary dependencies require Planner escalation; Builder must not unilaterally expand scope.
- Contract tests are distinct from TDD reproduction tests: TDD reproduction is mandatory for validated review findings; contract tests are required when boundary behavior is added, changed, or insufficiently protected.

## Ephemeral Communication Buffers

In-flight exchanges between agents (such as orientation inquiries, alignment dialogues, or detailed review findings) are transient transport messages.
When terminal viewports or transport scrollbacks necessitate buffering long messages into files:
- Prefer transport-native communication for in-flight exchanges.
- If buffering is necessary, use an existing already-ignored scratch directory if available (for example, `tmp/<task-id>/`).
- Otherwise, use a unique temporary directory via `mktemp -d /tmp/agentplaybook-XXXXXX`.
- Strictly forbid modifying `.git/`, `.git/info/exclude`, or `.gitignore` solely for communication buffering.
- Never commit or track files in scratch directories.

## Event-Driven Transport Coordination

When coordinating across agents in Herdr, follow a zero-poll policy: never use `sleep` loops or periodic screen polling. Use transport-native lifecycle events instead to preserve context tokens and avoid repetitive TUI captures. Consult the relevant transport skill for exact CLI syntax and options.

## Living AGENTS.md and Single-Writer Principle

`AgentPlaybook` adopts a single, living `AGENTS.md` at the repository root as an "Agent-Facing README" for instant $O(1)$ session bootstrapping across 5 canonical sections:
1. **Architectural Topology & Jurisdictions**: Repository tier, external interfaces, and jurisdictional boundaries (strict separation of concerns).
2. **Global Operational Invariants**: Project-level invariants (e.g., non-interactive execution guards, Conventional Commits, branch conventions).
3. **Builder Precautions & Gotchas**: Toolchain quirks, compiler limitations, and test runner constraints.
4. **Reviewer Precautions & Checklist**: Public verification guidelines and regression checkpoints (strictly excluding confidential test secrets).
5. **Active State & In-Flight Context**: Pre-commit baseline observation with mandatory provenance (`Observed-At: <UTC timestamp> @ <base-revision-id>`), dirty status, recent milestones, and next pickup item.

- **Single-Writer Principle**: Planner is the sole author and curator of `AGENTS.md`. Builder and Reviewer are strictly prohibited from editing it directly.
- **Ground Truth Revalidation**: Active State provides orienting context, never a substitute for live repository ground truth. Cold-starting Planners must execute fresh VCS status and log inspection commands to revalidate mutable facts before planning or executing tasks.
- **Blind-Barrier Check**: Reviewer conducts a narrow visibility check on `AGENTS.md` during the commit flow to ensure no confidential review criteria or hidden fixtures leak (`BARRIER_LEAK` returns to Planner for redaction).

### Language & Conciseness Standard

`AGENTS.md` must be authored in concise US English (`en-US`) and remain pure ASCII by default. A non-ASCII domain term is permitted only with an explicit adjacent inline rationale, such as `(domain term: concise rationale)`; this is the inline exception contract. Planner keeps entries concise and high-signal, excluding unnecessary verbosity, redundant explanations, and unverified operational narratives. Reviewer audits the file during the commit flow for unauthorized non-ASCII text lacking inline justification, excessive verbosity or bloat, and confidential barrier leakage.

### Telegraphic Memory & Inter-Agent Communication

`AGENTS.md` is machine-facing operational memory and a token-dense cache for LLM agents, not human-facing prose. Use concise fragments; drop articles, filler words, conversational pleasantries, and redundant grammar while preserving exact technical terms and code symbols. Inter-agent communication requires telegraphic (caveman) compression for all exchanges (including [Planner], [Reviewer], [Builder], [Scout], or [<role>] prefixed messages), omitting conversational filler and pleasantries in favor of compact structured technical fragments. This formatting is an endogenous specification only and introduces no new runtime packages, binaries, or dependencies.

### Role-Tiered Context Lifecycle & Compaction Governance
- Reviewer operates on high-tier reasoning models; after completing a commit flow, self-evaluate context utilization and compact or clean up when usage exceeds 50% or accumulated review history bloats operating costs.
- Builder operates as a stateless, disposable worker; replace bloated or quota-limited sessions with a fresh instance from the approved `<slug>.plan.md` rather than spending tokens on compaction.
- Planner retains orchestration history, curates living memory, and anchors transitions across ephemeral role lifecycles.

## Governed 9-Step Commit Flow (`flow commit`)

Version Control Governance is executed through a conceptual, evidence-based commit pipeline:
1. **Approval Baseline**: Confirm independent review approval (`REVIEW_PASS`) and establish candidate baseline.
2. **User Intent Gate**: Await explicit user commit request to initiate persistence.
3. **Caveats Query**: Query operational caveats from Builder and Reviewer (Reviewer reports only public, independently observable operational facts).
4. **Living Memory Update**: Planner updates `AGENTS.md` with synthesized caveats and fresh `Observed-At` provenance metadata.
5. **Visibility & Barrier Check**: Reviewer verifies `AGENTS.md` does not leak confidential review secrets (`AGENTS_REVIEW_PASS` -> 6; `BARRIER_LEAK` -> 4).
6. **Candidate Stabilization & Secret Scan**: Stabilize candidate snapshot identity, verify candidate equivalence, and execute fail-closed secret scan (`SCAN_CLEAN` -> 7; `SCAN_FAILED` -> 4).
7. **Commit Authorization Gate**: Present finalized candidate diff and Conventional Commit message, awaiting explicit human commit authorization (`AUTHORIZATION_GRANTED` -> 8; `AUTHORIZATION_DENIED` -> 2 awaiting renewed user intent).
8. **Local Revision Sealing**: Verify unchanged candidate identity and seal revision locally under Planner VCS governance (`PUBLICATION_AUTHORIZED` -> 9; else conclude locally).
9. **Remote Publication**: Publish sealed revision to remote repository under explicit publication authorization.

- **Commit & Publication Authority Separation**: Commit authorization permits local revision sealing only; remote publication requires separate, explicit human authorization.
- **Fail-Closed Intent Recovery**: Upon `AUTHORIZATION_DENIED`, Planner must return to Step 2 to await renewed user intent; autonomous re-drafting is strictly forbidden.

## Builder Diff Handoff and VCS Governance

Version Control Governance is exclusively owned by Planner; Builder delivers verified working copy diffs.

- Builder owns implementation craft: produce a minimal, reviewable working copy diff with reproduction and green unit tests, then hand off the verified diff and test logs to Planner.
- Builder must not execute VCS commit commands, modify commit history, or alter branch/revision pointers; hand off verified working copy diffs to Planner for VCS governance.
- Planner owns VCS history and revision progression: inspect the working copy diff against declared in-scope boundaries, execute the 9-Step Governed Commit Flow, and seal revisions under Planner VCS governance delegating to the active VCS/policy mechanism capabilities.
- These governance invariants are VCS-neutral: Builder delivers verified working copy diffs, and Planner owns commits and revision progression regardless of backend.
- Jujutsu is recommended for multi-agent collaboration because native change stacking supports isolated working-copy revisions, delegating command mechanics to the dedicated Jujutsu skill. Git is fully supported when the same Builder handoff and Planner governance are maintained.

## Discovery

Bare invocations are discovery-friendly and print concise catalogs with exit status 0:
- `agentplaybook`: overview manual
- `agentplaybook role`: list participant roles
- `agentplaybook flow`: list workflow procedures
- `agentplaybook artifact`: list document and message contracts
- `agentplaybook rule`: list rule commands
