# Implementation Plan: Workflow Collaboration Manual CLI

This plan defines the step-by-step implementation of the `workflow` Collaboration Manual CLI.
Each top-level checklist item corresponds to an atomic implementation unit mapped to a single `jj` commit.
Feature commits bundle their corresponding unit tests to ensure independent verifiability.

Commit message format:
```text
<type>(<scope>): <summary>

- detail 1
- detail 2
```

---

## Checklists & Commit Map

- [x] feat(data): add embedded knowledge JSON datasets and Go embed bundle
  - [x] Add `config.json` defining supported languages, transport (`herdr`), and prefix templates
  - [x] Add `roles.json` defining `planner`, `builder`, and `reviewer` identities and boundaries
  - [x] Add `flows.json` defining `init`, `plan`, `build`, and `review` 4-stage workflows with semantic condition tokens
  - [x] Add `artifacts.json` defining contracts for `repo-summary`, `build-plan`, `review-plan`, and `review-findings`
  - [x] Add `rules.json` defining `anti-cheating`, `mandatory-alignment`, `tdd-reproduction`, and `atomic-change-units`
  - [x] Create `internal/data/embed.go` using `//go:embed` to package all 5 JSON files into the Go binary

- [x] feat(knowledge): implement data models, loader, and 5-layer cross-file validator
  - [x] Implement Go model structs in `internal/knowledge/model.go` (`Config`, `RoleDefinition`, `Flow`, `FlowStep`, `Artifact`, `Rule`, `Knowledge`)
  - [x] Implement JSON decoding and map indexing in `internal/knowledge/load.go`
  - [x] Implement cross-file semantic validation in `internal/knowledge/validate.go` using `errors.Join` (validating roles, flows, rules, and artifacts)
  - [x] Implement query lookup helpers in `internal/knowledge/query.go`
  - [x] Add unit tests in `internal/knowledge/knowledge_test.go` verifying loading and validation

- [x] feat(cli): implement root command, entrypoint, and global flag handlers
  - [x] Initialize Cobra root command in `internal/cli/root.go` with `SilenceUsage: true` and `SilenceErrors: true`
  - [x] Create `main.go` entrypoint connecting to `cli.Execute()`
  - [x] Implement mutually-exclusive `--language`, `--transport`, and `--version` flags
  - [x] Implement output formatting helpers in `internal/cli/output.go` (structured JSON for data queries, clean text for discovery)
  - [x] Ensure bare `workflow` invocation outputs a concise discovery catalog with exit status 0
  - [x] Add unit tests in `internal/cli/root_test.go`

- [x] feat(cli): implement role query command with selector flags and tests
  - [x] Implement bare `workflow role` discovery catalog listing available roles and summaries
  - [x] Implement `workflow role <name>` full JSON role definition retrieval
  - [x] Implement mutually-exclusive selector flags (`--description`, `--responsibility`, `--boundary`, `--communication`)
  - [x] Add error handling and exit status 1 for unknown role names
  - [x] Add unit tests in `internal/cli/role_test.go`

- [x] feat(cli): implement flow query command with step isolation and tests
  - [x] Implement bare `workflow flow` discovery catalog listing the 4 available flows and summaries
  - [x] Implement `workflow flow <name>` full JSON flow retrieval
  - [x] Implement `--step <index>` flag to retrieve a single isolated `FlowStep`
  - [x] Add error handling for unknown flows and invalid/out-of-range step indices
  - [x] Add unit tests in `internal/cli/flow_test.go`

- [x] feat(cli): implement artifact contract query command and tests
  - [x] Implement bare `workflow artifact` discovery catalog listing all 4 artifacts, owners, and types
  - [x] Implement `workflow artifact <name>` retrieval displaying path, visibility, and required sections/fields
  - [x] Add error handling for unknown artifact names
  - [x] Add unit tests in `internal/cli/artifact_test.go`

- [ ] feat(cli): implement rule query command with explain support and tests
  - [ ] Implement bare `workflow rule` discovery catalog
  - [ ] Implement `workflow rule list` command displaying rule IDs, titles, categories, and summaries
  - [ ] Implement `workflow rule explain <id>...` command displaying full details for one or more rule IDs
  - [ ] Add error handling for unknown rule IDs
  - [ ] Add unit tests in `internal/cli/rule_test.go`

- [ ] feat(scripts): implement developer wrapper script with caching and dev overrides
  - [ ] Add `scripts/VERSION` specifying CLI version
  - [ ] Add `scripts/run-workflow.sh` supporting cache directory management (`~/.cache/workflow/<version>`) and `WORKFLOW_DEV=1` local dev override
  - [ ] Ensure executable permissions (`chmod +x scripts/run-workflow.sh`)
  - [ ] Verify local compilation via `go build -o bin/workflow main.go`

- [ ] test(cli): add golden integration tests across the complete CLI command tree
  - [ ] Implement golden tests in `internal/cli/golden_test.go` verifying output stability
  - [ ] Verify discovery commands consistently return exit status 0
  - [ ] Verify invalid queries return exit status 1 without usage dumps
  - [ ] Verify full test suite passes via `go test ./...`

- [ ] docs(skill): rewrite SKILL.md into thin progressive disclosure protocol guide
  - [ ] Rewrite root `SKILL.md` into a role-neutral ~70-line protocol guide
  - [ ] Document core philosophy: read-only collaboration manual, zero repository side-effects
  - [ ] Document on-demand query triggers (query role on identity setup, flow on procedure entry, artifact before authoring, rules on edge cases)
  - [ ] Document wrapper script invocation examples for Linux, macOS, and Windows
