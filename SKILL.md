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
   When establishing your participant identity or checking your allowed boundaries and responsibilities:
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

## Builder Diff Handoff and VCS Governance

Version Control Governance is exclusively owned by Planner; Builder delivers verified working copy diffs.

- Builder owns implementation craft: produce a minimal, reviewable working copy diff with reproduction and green unit tests, then hand off the verified diff and test logs to Planner.
- Builder must not execute VCS commit commands, modify commit history, or alter branch/revision pointers; hand off verified working copy diffs to Planner for VCS governance.
- Planner owns VCS history and revision progression: inspect the working copy diff against declared in-scope boundaries, create and seal the atomic Conventional Commit, and advance to the next revision.
- Git: Planner stages in-scope files and runs `git commit -m '...'` (which advances the active branch).
- Jujutsu: Planner describes the finalized revision with `jj describe -m '...'` and opens the next revision with `jj new`. Moving bookmarks (for example, `jj bookmark set <name> -r @-`) is optional on intermediate steps and can be deferred until the milestone or task is validated.

## Discovery

Bare invocations are discovery-friendly and print concise catalogs with exit status 0:
- `agentplaybook`: overview manual
- `agentplaybook role`: list participant roles
- `agentplaybook flow`: list workflow procedures
- `agentplaybook artifact`: list document and message contracts
- `agentplaybook rule`: list rule commands
