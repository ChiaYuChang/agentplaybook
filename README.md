# AgentPlaybook

[![Go Reference](https://pkg.go.dev/badge/github.com/ChiaYuChang/agentplaybook.svg)](https://pkg.go.dev/github.com/ChiaYuChang/agentplaybook)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

An on-demand, read-only collaboration manual and protocol reference for multi-agent coding workflows.

Instead of stuffing massive, static role prompts into every agent turn—wasting context tokens and risking cognitive drift—**AgentPlaybook** models workflow knowledge into five orthogonal domains and provides a token-efficient CLI for on-demand progressive disclosure.

---

## Key Features

- **Read-Only Collaboration Manual**: Zero side-effects on your target repository. No orchestrator runtime daemon, no external state store.
- **5 Orthogonal Knowledge Domains**:
  - `role`: Durable participant identities, boundaries, and communication targets (`planner`, `builder`, `reviewer`).
  - `flow`: Deterministic multi-agent procedures with semantic condition transitions (`init`, `plan`, `build`, `review`).
  - `artifact`: Document and message contracts specifying required sections and visibility boundaries (`repo-summary`, `build-plan`, `review-plan`, `review-findings`).
  - `rule`: Concrete operational policies and invariants (`anti-cheating`, `mandatory-alignment`, `atomic-change-units`, `tdd-reproduction`).
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
```bash
# Full role definition (JSON)
agentplaybook role builder

# Query specific sections
agentplaybook role builder --responsibility
agentplaybook role builder --boundary
agentplaybook role builder --communication
```

### 3. Inspecting Flows & Step SOPs
```bash
# Full flow definition
agentplaybook flow init

# Isolated single step query
agentplaybook flow build --step 2
```

### 4. Inspecting Artifact Contracts
```bash
agentplaybook artifact build-plan
agentplaybook artifact review-findings
```

### 5. Inspecting Behavioral Rules
```bash
# List all rule IDs and summaries
agentplaybook rule list

# Explain specific rule IDs
agentplaybook rule explain atomic-change-units
agentplaybook rule explain anti-cheating mandatory-alignment
```

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
