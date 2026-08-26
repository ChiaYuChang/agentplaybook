---
name: workflow
description: Use the Workflow Collaboration Manual CLI to coordinate multi-agent workflows, retrieve role boundaries, procedural flows, artifact contracts, and behavioral rules.
---

# Workflow Collaboration Manual CLI

The Workflow CLI is a read-only informational manual and reference playbook for multi-agent collaboration.
Use it as the definitive source of truth for participant roles, end-to-end orchestration flows, document contracts, and operational rules.

Commands:

- Linux or macOS: `sh "<skill-dir>/scripts/run-workflow.sh"`
- Local development: `WORKFLOW_DEV=1 sh "<skill-dir>/scripts/run-workflow.sh"`

## Core Philosophy

The Workflow CLI is an informational collaboration manual. It provides on-demand reference for roles, flows, artifact contracts, and rules.
The manual queries perform no workflow mutations, track no live state, spawn no agents, and execute no transport calls. It has zero side-effects on the target repository.
Interpret the retrieved guidance and perform the work using your own reasoning and tools.

## Progressive Disclosure Protocol

Do not query every knowledge domain on every turn. Query only the specific domain required for your immediate context:

1. **Role & Identity**:
   When establishing your participant identity or checking your allowed boundaries and responsibilities:
   ```sh
   sh "<skill-dir>/scripts/run-workflow.sh" role <role-name>
   # Selectors: --description, --responsibility, --boundary, --communication
   ```

2. **Flow & Procedures**:
   When entering a collaboration phase or determining the next step in a sequence:
   ```sh
   sh "<skill-dir>/scripts/run-workflow.sh" flow <flow-name>
   # Query single step: flow <flow-name> --step <index>
   ```

3. **Artifact Contracts**:
   Before authoring, reviewing, or exchanging persistent plans, summaries, or structured findings:
   ```sh
   sh "<skill-dir>/scripts/run-workflow.sh" artifact <artifact-name>
   ```

4. **Rules & Protocols**:
   When encountering specific boundary situations, test reproduction protocols, or invariant checks:
   ```sh
   sh "<skill-dir>/scripts/run-workflow.sh" rule list
   sh "<skill-dir>/scripts/run-workflow.sh" rule explain <rule-id>...
   ```

## Discovery

Bare invocations are discovery-friendly and print concise catalogs with exit status 0:
- `workflow`: overview manual
- `workflow role`: list participant roles
- `workflow flow`: list workflow procedures
- `workflow artifact`: list document and message contracts
- `workflow rule`: list rule commands