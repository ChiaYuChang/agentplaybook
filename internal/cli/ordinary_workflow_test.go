package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ChiaYuChang/agentplaybook/internal/cli"
	"github.com/ChiaYuChang/agentplaybook/internal/knowledge"
	versioninfo "github.com/ChiaYuChang/agentplaybook/internal/version"
)

func ordinaryExec(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := cli.Execute(args, &stdout, &stderr, "v0.1.0")
	return stdout.Bytes(), err
}

func ordinaryFlow(t *testing.T, name string) knowledge.Flow {
	t.Helper()
	out, err := ordinaryExec(t, "flow", name)
	if err != nil {
		t.Fatalf("flow %s failed: %v", name, err)
	}
	var f knowledge.Flow
	if err := json.Unmarshal(out, &f); err != nil {
		t.Fatalf("decode flow %s: %v", name, err)
	}
	return f
}

func ordinaryRole(t *testing.T, name string) knowledge.RoleDefinition {
	t.Helper()
	out, err := ordinaryExec(t, "role", name)
	if err != nil {
		t.Fatalf("role %s failed: %v", name, err)
	}
	var r knowledge.RoleDefinition
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("decode role %s: %v", name, err)
	}
	return r
}

func ordinaryArtifact(t *testing.T, name string) knowledge.Artifact {
	t.Helper()
	out, err := ordinaryExec(t, "artifact", name)
	if err != nil {
		t.Fatalf("artifact %s failed: %v", name, err)
	}
	var a knowledge.Artifact
	if err := json.Unmarshal(out, &a); err != nil {
		t.Fatalf("decode artifact %s: %v", name, err)
	}
	return a
}

func ordinaryRule(t *testing.T, id string) knowledge.Rule {
	t.Helper()
	out, err := ordinaryExec(t, "rule", "explain", id)
	if err != nil {
		t.Fatalf("rule explain %s failed: %v", id, err)
	}
	var rules []knowledge.Rule
	if err := json.Unmarshal(out, &rules); err != nil || len(rules) != 1 {
		t.Fatalf("decode rule %s: err=%v rules=%d", id, err, len(rules))
	}
	return rules[0]
}

func stepByIndex(f knowledge.Flow, idx int) (knowledge.FlowStep, bool) {
	for _, s := range f.Steps {
		if s.Index == idx {
			return s, true
		}
	}
	return knowledge.FlowStep{}, false
}

func stepConditions(s knowledge.FlowStep) map[string]int {
	m := make(map[string]int, len(s.Conditions))
	for _, c := range s.Conditions {
		m[c.When] = c.Then
	}
	return m
}

func flowCorpus(f knowledge.Flow) string {
	var b strings.Builder
	b.WriteString(f.Description)
	for _, s := range f.Steps {
		b.WriteString("\n")
		b.WriteString(s.Action)
		for _, c := range s.Conditions {
			b.WriteString("\n")
			b.WriteString(c.When)
		}
	}
	return b.String()
}

// reachable returns step indices reachable from start without following any
// condition whose trigger is in banned. Unconditional non-terminal steps fall
// through to Index+1.
func reachable(f knowledge.Flow, start int, banned ...string) map[int]bool {
	seen := map[int]bool{}
	var walk func(int)
	walk = func(idx int) {
		if seen[idx] {
			return
		}
		s, ok := stepByIndex(f, idx)
		if !ok {
			return
		}
		seen[idx] = true
		if s.Terminal {
			return
		}
		if len(s.Conditions) == 0 {
			walk(idx + 1)
			return
		}
		for _, c := range s.Conditions {
			if slices.Contains(banned, c.When) {
				continue
			}
			walk(c.Then)
		}
	}
	walk(start)
	return seen
}

func maxStepIndex(f knowledge.Flow) int {
	m := 0
	for _, s := range f.Steps {
		if s.Index > m {
			m = s.Index
		}
	}
	return m
}

// PLAN-01: one substantive Internal Reviewer approval and published TDD cases.
func TestOrdinary_PLAN01_SingleInternalPlanApproval(t *testing.T) {
	t.Parallel()
	f := ordinaryFlow(t, "plan")

	passGates := 0
	for _, s := range f.Steps {
		for _, c := range s.Conditions {
			if strings.HasSuffix(c.When, "_PASS") {
				passGates++
				if s.Actor != knowledge.RoleReviewer || c.When != "PLAN_REVIEW_PASS" {
					t.Errorf("unexpected approval gate %q at step %d actor %q", c.When, s.Index, s.Actor)
				}
			}
		}
	}
	if passGates != 1 {
		t.Errorf("ordinary plan flow must expose exactly one approval gate, got %d", passGates)
	}

	corpus := flowCorpus(f)
	for _, forbidden := range []string{"CONTRACT_REVIEW_PASS", "CONTRACT_REVIEW_REJECT", "CONTRACT_ONLY_REVISION", "SHARED_CONTRACT"} {
		if strings.Contains(corpus, forbidden) {
			t.Errorf("ordinary plan flow must not stack a separate contract gate (%s)", forbidden)
		}
	}
	for _, required := range []string{"Given/When/Then", "affected interaction contracts"} {
		if !strings.Contains(strings.ToLower(corpus), strings.ToLower(required)) {
			t.Errorf("ordinary plan flow must require %q", required)
		}
	}

	// Planner can reach the single review without any consultation.
	direct := reachable(f, 1, "CONSULTATION_REQUESTED")
	var reviewIdx int
	for _, s := range f.Steps {
		if _, ok := stepConditions(s)["PLAN_REVIEW_PASS"]; ok {
			reviewIdx = s.Index
		}
	}
	if reviewIdx == 0 || !direct[reviewIdx] {
		t.Fatalf("plan review (step %d) must be reachable without consultation: %v", reviewIdx, direct)
	}
	for idx := range direct {
		s, _ := stepByIndex(f, idx)
		if strings.Contains(strings.ToLower(s.Action), "consultation advice") {
			t.Errorf("mandatory path visits consultation step %d", idx)
		}
	}

	// PLAN_REVIEW_PASS directly authorizes finalization/Builder routing.
	pass := stepConditions(mustStep(t, f, reviewIdx))["PLAN_REVIEW_PASS"]
	finalize := mustStep(t, f, pass)
	if !strings.Contains(finalize.Action, "PLAN_REVIEW_PASS") || strings.Contains(finalize.Action, "CONTRACT") || !strings.Contains(finalize.Action, "build flow") {
		t.Errorf("PLAN_REVIEW_PASS must finalize and route to build flow without a second gate, got: %s", finalize.Action)
	}

	// Build handoff requires only the single approval.
	build := ordinaryFlow(t, "build")
	if !strings.Contains(build.Steps[0].Action, "PLAN_REVIEW_PASS") || strings.Contains(build.Steps[0].Action, "CONTRACT_REVIEW_PASS") {
		t.Errorf("build handoff must require PLAN_REVIEW_PASS only, got: %s", build.Steps[0].Action)
	}
}

func mustStep(t *testing.T, f knowledge.Flow, idx int) knowledge.FlowStep {
	t.Helper()
	s, ok := stepByIndex(f, idx)
	if !ok {
		t.Fatalf("flow %s has no step %d", f.Name, idx)
	}
	return s
}

// PLAN-02: optional consultation gives advice only and never approves Builder work.
func TestOrdinary_PLAN02_OptionalConsultationNeverApproves(t *testing.T) {
	t.Parallel()
	f := ordinaryFlow(t, "plan")

	first := mustStep(t, f, 1)
	conds := stepConditions(first)
	consultIdx, ok := conds["CONSULTATION_REQUESTED"]
	if !ok {
		t.Fatalf("step 1 must offer optional CONSULTATION_REQUESTED, got %v", conds)
	}
	if _, ok := conds["DIRECT_PLANNING"]; !ok {
		t.Fatalf("step 1 must offer DIRECT_PLANNING, got %v", conds)
	}
	consult := mustStep(t, f, consultIdx)
	if consult.Actor != knowledge.RoleReviewer && consult.Actor != knowledge.RoleScout {
		t.Errorf("consultation actor must be an advisory peer, got %q", consult.Actor)
	}
	lower := strings.ToLower(consult.Action)
	for _, phrase := range []string{"not plan approval", "never authorizes builder"} {
		if !strings.Contains(lower, phrase) {
			t.Errorf("consultation must state %q, got: %s", phrase, consult.Action)
		}
	}
	for when := range stepConditions(consult) {
		if strings.HasSuffix(when, "_PASS") {
			t.Errorf("consultation must not be an approval gate (%s)", when)
		}
	}
	// Without PLAN_REVIEW_PASS the Builder-routing final step is unreachable from consultation.
	noPass := reachable(f, consultIdx, "PLAN_REVIEW_PASS")
	if noPass[maxStepIndex(f)] {
		t.Errorf("consultation must not reach Builder routing without PLAN_REVIEW_PASS: %v", noPass)
	}
}

// BUILD-01 and TDD-01: bounded autonomy, public scenarios drive RED/GREEN/refactor.
func TestOrdinary_BUILD01_TDD01_AutonomyAndPublicTDD(t *testing.T) {
	t.Parallel()
	f := ordinaryFlow(t, "build")

	corpus := flowCorpus(f)
	lowerCorpus := strings.ToLower(corpus)
	if strings.Contains(corpus, "ALIGNMENT_COMPLETE") || strings.Contains(lowerCorpus, "alignment") {
		t.Errorf("build flow must not impose a mandatory implementation alignment gate")
	}
	if !strings.Contains(f.Description, "PLAN_REVIEW_PASS") || strings.Contains(f.Description, "CONTRACT_REVIEW_PASS") {
		t.Errorf("build description must require only PLAN_REVIEW_PASS, got: %s", f.Description)
	}

	// Builder starts implementing immediately after blind handoff.
	second := mustStep(t, f, 2)
	if second.Actor != knowledge.RoleBuilder {
		t.Fatalf("builder must act directly after handoff, got step 2 actor %q", second.Actor)
	}
	var builderText strings.Builder
	for _, s := range f.Steps {
		if s.Actor == knowledge.RoleBuilder {
			builderText.WriteString(s.Action)
			builderText.WriteString("\n")
		}
	}
	bt := builderText.String()
	for _, phrase := range []string{"authorized module/behavior boundary", "explicit exclusions", "RED", "GREEN", "refactor", "intended assertion", "setup", "public", "protected shared or public contract", "static/specification evidence"} {
		if !strings.Contains(bt, phrase) {
			t.Errorf("builder steps must mention %q", phrase)
		}
	}
	// Escalation routes to Planner and back; it cannot bypass verification.
	esc := stepConditions(second)["ESCALATION_REQUIRED"]
	if esc == 0 {
		t.Fatalf("builder step must expose ESCALATION_REQUIRED, got %v", stepConditions(second))
	}
	if mustStep(t, f, esc).Actor != knowledge.RolePlanner {
		t.Errorf("escalation must route to planner")
	}
	if !strings.Contains(strings.ToLower(mustStep(t, f, esc).Action), "returns to planner") && !strings.Contains(strings.ToLower(mustStep(t, f, esc).Action), "new requirement") {
		t.Errorf("escalation step must route new requirements through Planner, got: %s", mustStep(t, f, esc).Action)
	}

	// Blind barrier retained at handoff.
	if !strings.Contains(f.Steps[0].Action, "withhold review-plan") {
		t.Errorf("handoff must withhold review-plan, got: %s", f.Steps[0].Action)
	}

	// Plan publishes the TDD cases before implementation.
	plan := ordinaryFlow(t, "plan")
	if !strings.Contains(flowCorpus(plan), "before implementation") {
		t.Errorf("plan flow must publish TDD scenarios before implementation")
	}
	bp := ordinaryArtifact(t, "build-plan")
	var tddSection *knowledge.ArtifactSection
	for i := range bp.Sections {
		if bp.Sections[i].Name == "TDD Scenarios" {
			tddSection = &bp.Sections[i]
		}
	}
	if tddSection == nil || !tddSection.Required || !strings.Contains(tddSection.Description, "Given/When/Then") {
		t.Errorf("build-plan must require a public 'TDD Scenarios' section with Given/When/Then, got %+v", tddSection)
	}
	if !slices.Contains(bp.Visibility, knowledge.RoleBuilder) {
		t.Errorf("Builder must see the public build-plan")
	}

	// Reviewer evaluates tests independently after submission; no hidden prerequisite.
	review := ordinaryFlow(t, "review")
	step2 := mustStep(t, review, 2).Action
	for _, phrase := range []string{"test adequacy", "independent counterexamples", "no advance secret fixture"} {
		if !strings.Contains(step2, phrase) {
			t.Errorf("review step 2 must mention %q, got: %s", phrase, step2)
		}
	}
	tdd := ordinaryRule(t, "tdd-reproduction")
	joined := strings.Join(tdd.Guidelines, "\n")
	for _, phrase := range []string{"public Given/When/Then", "no hidden", "returns to Planner"} {
		if !strings.Contains(joined, phrase) {
			t.Errorf("tdd-reproduction must state %q", phrase)
		}
	}

	// Role contracts: bounded autonomy replaces alignment dependence.
	builder := ordinaryRole(t, "builder")
	for _, b := range builder.Boundaries {
		if strings.Contains(b, "prior alignment") {
			t.Errorf("builder boundary still requires prior alignment: %q", b)
		}
	}
	if !strings.Contains(strings.Join(builder.Responsibilities, "\n"), "authorized module/behavior boundary") {
		t.Errorf("builder responsibilities must define bounded autonomy")
	}
	if _, err := ordinaryExec(t, "rule", "explain", "mandatory-alignment"); err == nil {
		t.Errorf("mandatory-alignment rule must be retired for ordinary work")
	}
	auto := ordinaryRule(t, "bounded-builder-autonomy")
	if !strings.Contains(strings.Join(auto.Guidelines, "\n"), "explicit exclusions") {
		t.Errorf("bounded-builder-autonomy must define exclusions: %v", auto.Guidelines)
	}
}

// REVIEW-01: internal approval is honest about independence and requires evidence.
func TestOrdinary_REVIEW01_InternalReviewHonesty(t *testing.T) {
	t.Parallel()
	review := ordinaryFlow(t, "review")
	step2 := mustStep(t, review, 2).Action
	for _, phrase := range []string{"internal technical approval", "not an independent external assessment", "concrete evidence", "preference", "Delta-review"} {
		if !strings.Contains(step2, phrase) {
			t.Errorf("review step 2 must mention %q, got: %s", phrase, step2)
		}
	}
	// Existing conditions remain intact.
	conds := stepConditions(mustStep(t, review, 2))
	if conds["REVIEW_PASS"] != 6 || conds["IMPLEMENTATION_REVIEW_PASS"] != 6 || conds["FINDINGS_REPORTED"] != 3 || conds["BASELINE_STALE"] != 8 {
		t.Errorf("review step 2 transitions changed unexpectedly: %v", conds)
	}

	res := ordinaryArtifact(t, "review-resolution")
	var outcome string
	for _, s := range res.Sections {
		if s.Name == "Outcome" {
			outcome = s.Description
		}
	}
	for _, phrase := range []string{"internal technical", "external review status", "not externally assessed"} {
		if !strings.Contains(outcome, phrase) {
			t.Errorf("review-resolution Outcome must state %q, got: %s", phrase, outcome)
		}
	}

	reviewer := ordinaryRole(t, "reviewer")
	if !strings.Contains(reviewer.Description, "Internal Reviewer") {
		t.Errorf("reviewer must be identified as the Internal Reviewer, got: %s", reviewer.Description)
	}
	for _, r := range reviewer.Boundaries {
		if strings.Contains(r, "require separate CONTRACT_REVIEW_PASS") && !strings.Contains(r, "legacy") {
			t.Errorf("ordinary reviewer boundary still demands separate contract pass: %s", r)
		}
	}
}

// EXT-01: no automatic external gate; user request and publication stay separate.
func TestOrdinary_EXT01_NoAutomaticExternalGate(t *testing.T) {
	t.Parallel()

	out, err := ordinaryExec(t, "flow")
	if err != nil {
		t.Fatalf("flow listing failed: %v", err)
	}
	if !json.Valid(out) && len(out) == 0 {
		t.Fatalf("empty flow listing")
	}
	for _, name := range []string{"init", "plan", "blueprint", "build", "review", "commit", "session-handoff", "cartography", "navigator-cartography", "e2e"} {
		f := ordinaryFlow(t, name)
		for _, s := range f.Steps {
			if s.Actor == knowledge.RoleExternalReviewer {
				t.Errorf("flow %s step %d must not gate on external-reviewer", name, s.Index)
			}
			for _, c := range s.Conditions {
				if strings.Contains(strings.ToUpper(c.When), "EXTERNAL") {
					t.Errorf("flow %s step %d has automatic external condition %q", name, s.Index, c.When)
				}
			}
		}
	}

	commit := ordinaryFlow(t, "commit")
	corpus := strings.ToLower(flowCorpus(commit))
	for _, phrase := range []string{"user-requested external review", "not externally assessed", "never substitutes for publication authorization", "fixed candidate"} {
		if !strings.Contains(corpus, phrase) {
			t.Errorf("commit flow must state %q", phrase)
		}
	}
	// Publication still requires the explicit PUBLICATION_AUTHORIZED transition.
	if reachable(commit, 1, "PUBLICATION_AUTHORIZED")[9] {
		t.Errorf("step 9 publication must be unreachable without PUBLICATION_AUTHORIZED")
	}

	rule := ordinaryRule(t, "user-requested-external-review")
	joined := strings.ToLower(strings.Join(rule.Guidelines, "\n") + rule.Details + rule.Summary)
	for _, phrase := range []string{"only on explicit user request", "fixed candidate", "does not authorize remote push", "after local seal", "not externally assessed", "sanitized"} {
		if !strings.Contains(joined, phrase) {
			t.Errorf("user-requested-external-review must state %q", phrase)
		}
	}

	acc := ordinaryRule(t, "acceptance-publication-authority")
	if !strings.Contains(strings.Join(acc.Guidelines, "\n"), "external review") {
		t.Errorf("acceptance-publication-authority must keep external review separate from publication")
	}
}

// EXT-02: role-level boundaries are visible through the public CLI.
func TestOrdinary_EXT02_RoleBoundariesViaCLI(t *testing.T) {
	t.Parallel()

	ext := ordinaryRole(t, "external-reviewer")
	if ext.Category != "core" || len(ext.Communication.Targets) != 1 || ext.Communication.Targets[0] != knowledge.RolePlanner {
		t.Errorf("external-reviewer must be a core role targeting only planner: %+v", ext)
	}
	desc := strings.ToLower(ext.Description)
	if !strings.Contains(desc, "user-requested") || !strings.Contains(desc, "fresh") {
		t.Errorf("external-reviewer description must state user-requested fresh session, got: %s", ext.Description)
	}
	boundaries := strings.Join(ext.Boundaries, "\n")
	for _, phrase := range []string{"review-plan", "review-findings", "Builder", "fixed candidate", "AGENTS.md"} {
		if !strings.Contains(boundaries, phrase) {
			t.Errorf("external-reviewer boundaries must mention %q", phrase)
		}
	}

	planner := ordinaryRole(t, "planner")
	if len(planner.Communication.Targets) != 7 || !slices.Contains(planner.Communication.Targets, knowledge.RoleExternalReviewer) {
		t.Errorf("planner targets must add external-reviewer: %v", planner.Communication.Targets)
	}
	pr := strings.Join(planner.Responsibilities, "\n") + strings.Join(planner.Boundaries, "\n")
	for _, phrase := range []string{"external-review-report", "only on explicit user request"} {
		if !strings.Contains(pr, phrase) {
			t.Errorf("planner role must mention %q", phrase)
		}
	}

	for _, private := range []string{"review-plan", "review-findings"} {
		a := ordinaryArtifact(t, private)
		if slices.Contains(a.Visibility, knowledge.RoleExternalReviewer) {
			t.Errorf("%s leaked to external-reviewer: %v", private, a.Visibility)
		}
		if !slices.Contains(a.Visibility, knowledge.RoleReviewer) {
			t.Errorf("Internal Reviewer lost %s access: %v", private, a.Visibility)
		}
	}
	// Existing Internal Reviewer retains E2E participation and contract.
	e2e := ordinaryFlow(t, "e2e")
	if len(e2e.Steps) != 12 || mustStep(t, e2e, 8).Actor != knowledge.RoleReviewer {
		t.Errorf("e2e flow must retain Internal Reviewer participation: %d steps", len(e2e.Steps))
	}
	reviewer := ordinaryRole(t, "reviewer")
	if reviewer.Name != knowledge.RoleReviewer || len(reviewer.Communication.Targets) != 1 {
		t.Errorf("reviewer identity changed: %+v", reviewer)
	}
}

// EXT-03: the report is a Planner-only evidence channel and never authorizes publication.
func TestOrdinary_EXT03_ExternalReviewReportViaCLI(t *testing.T) {
	t.Parallel()

	a := ordinaryArtifact(t, "external-review-report")
	if a.Owner != knowledge.RoleExternalReviewer || a.Type != "message" {
		t.Fatalf("unexpected report metadata: %+v", a)
	}
	if len(a.Visibility) != 2 || a.Visibility[0] != knowledge.RolePlanner || a.Visibility[1] != knowledge.RoleExternalReviewer {
		t.Errorf("report visibility must be [planner external-reviewer], got %v", a.Visibility)
	}
	fields := map[string]knowledge.ArtifactField{}
	for _, f := range a.Fields {
		fields[f.Name] = f
	}
	for _, name := range []string{"candidate_ref", "scope", "outcome", "evidence", "verification_assessment"} {
		if f, ok := fields[name]; !ok || !f.Required {
			t.Errorf("field %q must be required, got %+v", name, f)
		}
	}
	if f, ok := fields["findings"]; !ok || f.Required {
		t.Errorf("field findings must be optional, got %+v", f)
	}
	for _, token := range []string{"PASS", "FINDINGS", "INCOMPLETE"} {
		if !strings.Contains(fields["outcome"].Description, token) {
			t.Errorf("outcome must define %s, got: %s", token, fields["outcome"].Description)
		}
	}
	if !strings.Contains(fields["candidate_ref"].Description, "fixed") {
		t.Errorf("candidate_ref must identify a fixed revision, got: %s", fields["candidate_ref"].Description)
	}
	d := strings.ToLower(a.Description)
	for _, phrase := range []string{"does not authorize remote push", "planner", "sanitize"} {
		if !strings.Contains(d, phrase) {
			t.Errorf("report description must state %q, got: %s", phrase, a.Description)
		}
	}
}

// SHARED-01: ordinary shared-contract change is targeted; legacy Blueprint path is preserved and opt-in.
func TestOrdinary_SHARED01_TargetedInvalidationAndLegacyPreserved(t *testing.T) {
	t.Parallel()

	plan := ordinaryFlow(t, "plan")
	corpus := flowCorpus(plan)
	for _, phrase := range []string{"optional architectural Blueprint view", "legacy Blueprint flow", "opt-in", "invalidate only affected", "integration acceptance"} {
		if !strings.Contains(corpus, phrase) {
			t.Errorf("plan flow must state %q", phrase)
		}
	}
	if strings.Contains(corpus, "Hierarchical Blueprint and Single Plan") {
		t.Errorf("plan flow must not branch to the Blueprint lifecycle by size")
	}
	// No ordinary path into the legacy blueprint lifecycle by size alone.
	if strings.Contains(corpus, "route single-plan work to build flow or hierarchical work to blueprint flow") {
		t.Errorf("plan flow must not route hierarchical work to blueprint flow by default")
	}

	bpArtifact := ordinaryArtifact(t, "build-plan")
	var view *knowledge.ArtifactSection
	for i := range bpArtifact.Sections {
		if strings.HasPrefix(bpArtifact.Sections[i].Name, "Architectural View") {
			view = &bpArtifact.Sections[i]
		}
	}
	if view == nil || view.Required {
		t.Errorf("build-plan must expose an optional 'Architectural View' section, got %+v", view)
	}
	if _, err := ordinaryExec(t, "artifact", "blueprint-plan"); err != nil {
		t.Errorf("blueprint-plan artifact must remain: %v", err)
	}

	// Legacy blueprint invalidation semantics intact.
	bp := ordinaryFlow(t, "blueprint")
	if len(bp.Steps) != 17 {
		t.Fatalf("legacy blueprint must keep 17 steps, got %d", len(bp.Steps))
	}
	if !strings.Contains(mustStep(t, bp, 4).Action, "invalidates prior approvals for dependent sub-plans") {
		t.Errorf("legacy blueprint dependent-approval invalidation must remain")
	}
	if !strings.Contains(mustStep(t, bp, 5).Action, "Reassess every dependent sub-plan through Sub-Plan and contract gates") {
		t.Errorf("legacy blueprint reassessment must remain")
	}

	rule := ordinaryRule(t, "interaction-contract-gate")
	joined := rule.Summary + rule.Details + strings.Join(rule.Guidelines, "\n")
	for _, phrase := range []string{"ordinary", "legacy Blueprint", "single", "affected"} {
		if !strings.Contains(joined, phrase) {
			t.Errorf("interaction-contract-gate must distinguish ordinary vs legacy (%q)", phrase)
		}
	}

	review := ordinaryFlow(t, "review")
	final := mustStep(t, review, 6).Action
	for _, phrase := range []string{"coverage", "affected revision identities"} {
		if !strings.Contains(final, phrase) {
			t.Errorf("review resolution step must record %q, got: %s", phrase, final)
		}
	}
}

// E2E-01: required E2E blocks final PASS and local seal even while paused.
func TestOrdinary_E2E01_RequiredE2ERemainsBlocking(t *testing.T) {
	t.Parallel()

	review := ordinaryFlow(t, "review")
	final := mustStep(t, review, 6)
	if !final.Terminal {
		t.Errorf("review step 6 must remain terminal")
	}
	for _, phrase := range []string{"E2E_REQUIRED", "E2E_NOT_REQUIRED", "paused", "blocks final REVIEW_PASS", "never relabeled"} {
		if !strings.Contains(final.Action, phrase) {
			t.Errorf("review step 6 must state %q, got: %s", phrase, final.Action)
		}
	}
	commit := ordinaryFlow(t, "commit")
	first := mustStep(t, commit, 1).Action
	for _, phrase := range []string{"E2E_EVIDENCE_ADMITTED", "local seal"} {
		if !strings.Contains(first, phrase) {
			t.Errorf("commit step 1 must gate on %q, got: %s", phrase, first)
		}
	}
	// E2E admission semantics untouched.
	e2e := ordinaryFlow(t, "e2e")
	if !mustStep(t, e2e, 11).Terminal || !mustStep(t, e2e, 12).Terminal {
		t.Errorf("e2e terminal steps must be preserved")
	}
	for when, then := range stepConditions(mustStep(t, e2e, 6)) {
		_ = then
		if strings.Contains(when, "EXTERNAL") {
			t.Errorf("e2e admission must not gain external conditions")
		}
	}
}

// Strengthened public contract test: prove no Builder-routing finalization path from plan entry can avoid PLAN_REVIEW_PASS.
func TestOrdinary_PLAN01_NoPathToBuilderRoutingWithoutPlanReviewPass(t *testing.T) {
	t.Parallel()
	f := ordinaryFlow(t, "plan")

	// Identify terminal Builder-routing finalization step
	var finalStep int
	for _, s := range f.Steps {
		if s.Terminal || strings.Contains(strings.ToLower(s.Action), "build flow") {
			finalStep = s.Index
		}
	}
	if finalStep == 0 {
		t.Fatalf("plan flow has no terminal Builder-routing step")
	}

	// 1. Direct and consultation entry points cannot reach final step without PLAN_REVIEW_PASS
	for _, entry := range []struct {
		name  string
		start int
	}{
		{"flow_entry_step_1", 1},
		{"consultation_step_2", 2},
		{"direct_authoring_step_3", 3},
		{"submission_step_4", 4},
		{"review_evaluation_step_5", 5},
		{"planner_revision_step_6", 6},
	} {
		reachableWithoutPass := reachable(f, entry.start, "PLAN_REVIEW_PASS")
		if reachableWithoutPass[finalStep] {
			t.Errorf("path from %s (step %d) reaches final Builder-routing step %d without PLAN_REVIEW_PASS: %v",
				entry.name, entry.start, finalStep, reachableWithoutPass)
		}
	}

	// 2. Step 1 transitions: only CONSULTATION_REQUESTED -> 2 and DIRECT_PLANNING -> 3
	step1 := mustStep(t, f, 1)
	s1Conds := stepConditions(step1)
	if len(s1Conds) != 2 || s1Conds["CONSULTATION_REQUESTED"] != 2 || s1Conds["DIRECT_PLANNING"] != 3 {
		t.Errorf("unexpected step 1 transitions: %v", s1Conds)
	}

	// 3. Step 2 (consultation) transitions: only CONSULTATION_CONCLUDED -> 3
	step2 := mustStep(t, f, 2)
	s2Conds := stepConditions(step2)
	if len(s2Conds) != 1 || s2Conds["CONSULTATION_CONCLUDED"] != 3 {
		t.Errorf("unexpected step 2 transitions: %v", s2Conds)
	}

	// 4. Step 5 (review) transitions: only PLAN_REVIEW_PASS -> 7 and PLAN_REVIEW_REJECT -> 6
	step5 := mustStep(t, f, 5)
	s5Conds := stepConditions(step5)
	if len(s5Conds) != 2 || s5Conds["PLAN_REVIEW_PASS"] != finalStep || s5Conds["PLAN_REVIEW_REJECT"] != 6 {
		t.Errorf("unexpected step 5 transitions: %v", s5Conds)
	}

	// 5. Step 6 (revision after rejection) transitions: only PLANS_REVISED -> 4 (must resubmit, cannot jump to 5 or 7)
	step6 := mustStep(t, f, 6)
	s6Conds := stepConditions(step6)
	if len(s6Conds) != 1 || s6Conds["PLANS_REVISED"] != 4 {
		t.Errorf("unexpected step 6 transitions (must route to resubmission step 4): %v", s6Conds)
	}
}

// Template/docs must not describe contradictory ordinary workflows.
func TestOrdinary_TemplatesDescribeSingleInternalApproval(t *testing.T) {
	t.Parallel()

	std := cli.DefaultLivingMemoryTemplate()
	for _, forbidden := range []string{"Dual Formal Gates", "Two distinct validation gates", "separate Reviewer `CONTRACT_REVIEW_PASS`"} {
		if strings.Contains(std, forbidden) {
			t.Errorf("default template still describes superseded gates: %q", forbidden)
		}
	}
	for _, required := range []string{
		"external-reviewer",
		"PLAN_REVIEW_PASS",
		"E2E_EVIDENCE_ADMITTED",
		"external-review-report",
		"Given/When/Then",
		"module/behavior",
		"explicit exclusions",
		"legacy Blueprint",
		"user-requested external review",
		"does not require Blueprint or sub-plans",
		"In opt-in legacy Blueprint flows, sub-plans are authored Just-In-Time (JIT)",
	} {
		if !strings.Contains(std, required) {
			t.Errorf("default template missing %q", required)
		}
	}

	min := cli.MinimalLivingMemoryTemplate()
	if strings.Contains(min, "Dual Gates") {
		t.Errorf("minimal template still describes dual gates")
	}
	if strings.Contains(min, "(or legacy PLAN_PASS)") {
		t.Errorf("minimal template must use actual PLAN_REVIEW_PASS without legacy PLAN_PASS fallback")
	}
	for _, required := range []string{
		"external-reviewer",
		"PLAN_REVIEW_PASS",
		"IMPLEMENTATION_REVIEW_PASS",
		"REVIEW_PASS",
		"Given/When/Then",
		"module/behavior",
		"explicit exclusions",
		"legacy Blueprint",
		"external review",
	} {
		if !strings.Contains(min, required) {
			t.Errorf("minimal template missing %q", required)
		}
	}

	skillBytes, err := os.ReadFile("../../SKILL.md")
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	skillContent := string(skillBytes)
	if strings.Contains(skillContent, "`PLAN_REVIEW_PASS` and `SUBPLAN_PLAN_PASS` remain necessary but are not sufficient") {
		t.Errorf("SKILL.md still claims PLAN_REVIEW_PASS is not sufficient for Builder routing")
	}
	for _, required := range []string{
		"one substantive internal `PLAN_REVIEW_PASS` authorizes Builder routing",
		"ordinary shared-contract changes reassess affected consumers",
		"strictly reserved for opt-in legacy Blueprint",
	} {
		if !strings.Contains(skillContent, required) {
			t.Errorf("SKILL.md missing contract gate qualification: %q", required)
		}
	}
}

// Catalog role and artifact descriptions must qualify two-gate wording strictly to legacy Blueprint.
func TestOrdinary_CatalogRoleArtifactTwoGateWordingQualified(t *testing.T) {
	t.Parallel()

	// build-plan artifact description
	bp := ordinaryArtifact(t, "build-plan")
	if !strings.Contains(bp.Description, "PLAN_REVIEW_PASS") {
		t.Errorf("build-plan description must state PLAN_REVIEW_PASS: %s", bp.Description)
	}
	if strings.Contains(bp.Description, "both concept/method and interaction-contract review gates") && !strings.Contains(bp.Description, "legacy Blueprint") {
		t.Errorf("build-plan description must qualify dual gates to legacy Blueprint: %s", bp.Description)
	}

	// Planner responsibilities
	planner := ordinaryRole(t, "planner")
	plannerCorpus := strings.Join(planner.Responsibilities, "\n")
	if !strings.Contains(plannerCorpus, "PLAN_REVIEW_PASS") {
		t.Errorf("planner responsibilities must state PLAN_REVIEW_PASS")
	}
	if strings.Contains(plannerCorpus, "independent concept/method and contract gates") && !strings.Contains(plannerCorpus, "legacy Blueprint") {
		t.Errorf("planner responsibilities must qualify dual gates to legacy Blueprint: %s", plannerCorpus)
	}

	// Reviewer responsibilities and boundaries
	reviewer := ordinaryRole(t, "reviewer")
	reviewerResp := strings.Join(reviewer.Responsibilities, "\n")
	if strings.Contains(reviewerResp, "after the concept/method gate") && !strings.Contains(reviewerResp, "legacy Blueprint") {
		t.Errorf("reviewer responsibilities must qualify post-concept contract review to legacy Blueprint: %s", reviewerResp)
	}

	// interaction-contract-gate rule guidelines
	rule := ordinaryRule(t, "interaction-contract-gate")
	ruleCorpus := strings.Join(rule.Guidelines, "\n")
	if strings.Contains(ruleCorpus, "distinct contract-review approval before Builder handoff") && !strings.Contains(ruleCorpus, "legacy Blueprint") {
		t.Errorf("interaction-contract-gate rule must qualify distinct contract approval to legacy Blueprint: %s", ruleCorpus)
	}
}

// VERSION-01: canonical version is exactly v0.5.0 across version.Release(), CLI output, and init template header.
func TestOrdinary_VERSION01_ReleaseVersionV050(t *testing.T) {
	t.Parallel()

	if got := versioninfo.Release(); got != "v0.5.0" {
		t.Errorf("versioninfo.Release() = %q, want v0.5.0", got)
	}

	tmpl := cli.DefaultLivingMemoryTemplate()
	if !strings.Contains(tmpl, "AgentPlaybook v0.5.0 Living Memory Blueprint") {
		t.Errorf("default template header must reference v0.5.0, got: %s", tmpl[:200])
	}
}
