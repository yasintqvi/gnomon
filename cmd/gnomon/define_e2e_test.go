package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeDefineAgent is a stand-in Agent executable: it reads the prompt Gnomon wrote, appends it to
// $FAKE_PROMPT_LOG, and publishes a Define result whose outcome $FAKE_MODE chooses — or, for
// "crash", exits without any result, like an interrupted session.
const fakeDefineAgent = `#!/bin/sh
prompt=${1#Read and follow the instructions in }
cat "$prompt" >> "$FAKE_PROMPT_LOG"
printf '\n=====\n' >> "$FAKE_PROMPT_LOG"
result=$(grep -A1 'publish exactly one valid JSON result to this path:' "$prompt" | tail -1)
runid=$(sed -n 's/.*"run_id": "\([^"]*\)".*/\1/p' "$prompt" | head -1)
case "$FAKE_MODE" in
  crash) exit 3 ;;
  blocked) outcome=BLOCKED ;;
  *) outcome=READY_FOR_APPROVAL ;;
esac
printf '{"workflow":"specification-definition","run_id":"%s","payload":{"outcome":"%s"}}' "$runid" "$outcome" > "$result"
exec sleep 10
`

type defineEnv struct {
	t         *testing.T
	dir       string
	env       []string
	promptLog string
}

func newDefineEnv(t *testing.T) *defineEnv {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in Agent is a POSIX shell script")
	}
	dir, env := gitInitializedRepo(t)
	agent := filepath.Join(t.TempDir(), "fake-agent")
	if err := os.WriteFile(agent, []byte(fakeDefineAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "prompts.log")
	env = append(env, "GNOMON_CLAUDE_EXECUTABLE="+agent, "FAKE_PROMPT_LOG="+log, "NO_COLOR=1")
	if code, out := runGnomonOutput(t, dir, env, "init"); code != 0 {
		t.Fatalf("init: %s", out)
	}
	return &defineEnv{t: t, dir: dir, env: env, promptLog: log}
}

func (d *defineEnv) run(mode string, args ...string) (int, string) {
	d.t.Helper()
	env := append(append([]string{}, d.env...), "FAKE_MODE="+mode)
	return runGnomonOutput(d.t, d.dir, env, append(args, "--agent", "claude")...)
}

func (d *defineEnv) specFiles() []string {
	d.t.Helper()
	entries, _ := os.ReadDir(filepath.Join(d.dir, ".gnomon", "specifications"))
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "SPEC-") {
			names = append(names, e.Name())
		}
	}
	return names
}

func (d *defineEnv) lastPrompt() string {
	data, _ := os.ReadFile(d.promptLog)
	parts := strings.Split(strings.TrimSuffix(string(data), "\n=====\n"), "\n=====\n")
	return parts[len(parts)-1]
}

const reserveRequest = "Members can reserve a tool that is out. They wait in line for it."

// One command starts a new feature: the Draft is created, its identity shown, the request saved
// in it and handed to the Agent. Repeating the request continues the same Draft.
func TestDefine_NewRequest_CreatesDraftOnce(t *testing.T) {
	d := newDefineEnv(t)

	code, out := d.run("ready", "define", reserveRequest)
	if code != 0 {
		t.Fatalf("define exit %d:\n%s", code, out)
	}
	for _, want := range []string{"Created SPEC-001 — Members can reserve a tool that is out (Draft)", "Ready for approval"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in:\n%s", want, out)
		}
	}
	files := d.specFiles()
	if len(files) != 1 || files[0] != "SPEC-001-members-can-reserve-a-tool-that-is-out.md" {
		t.Fatalf("expected one new Specification, got %v", files)
	}
	content, _ := os.ReadFile(filepath.Join(d.dir, ".gnomon", "specifications", files[0]))
	if !strings.Contains(string(content), "> Request: "+reserveRequest) {
		t.Fatalf("expected the request saved in the Draft:\n%s", content)
	}
	if !strings.Contains(d.lastPrompt(), "The user's request for this run, in their own words:\n"+reserveRequest) {
		t.Fatalf("expected the request handed to the Agent:\n%s", d.lastPrompt())
	}

	code, out = d.run("ready", "define", reserveRequest)
	if code != 0 || !strings.Contains(out, "Continuing SPEC-001 — Members can reserve a tool that is out") {
		t.Fatalf("repeating the request must continue the same Draft, exit %d:\n%s", code, out)
	}
	if len(d.specFiles()) != 1 {
		t.Fatalf("no duplicate Specification may be created, got %v", d.specFiles())
	}
}

// A different request that would get the same title is refused instead of silently continuing
// someone else's Draft; --title separates them.
func TestDefine_SameTitleDifferentRequest_RefusedThenTitleSeparates(t *testing.T) {
	d := newDefineEnv(t)
	if code, out := d.run("ready", "define", "Late fees", "--title", "Late fees"); code != 0 {
		t.Fatalf("define: %s", out)
	}
	code, out := d.run("ready", "define", "Late fees: 2 euros per day", "--title", "Late fees")
	if code == 0 || !strings.Contains(out, "SPEC-001 — Late fees is a Draft started from a different request") {
		t.Fatalf("expected a refusal naming the existing Draft, exit %d:\n%s", code, out)
	}
	if len(d.specFiles()) != 1 {
		t.Fatalf("a refused request must not create a Specification: %v", d.specFiles())
	}
	if code, out := d.run("ready", "define", "Late fees: 2 euros per day", "--title", "Damage fees"); code != 0 || !strings.Contains(out, "Created SPEC-002 — Damage fees") {
		t.Fatalf("a different --title starts a separate Specification, exit %d:\n%s", code, out)
	}
}

// An existing Specification is defined by its identity, with any further text handed over as the
// change; nothing new is created. An unknown identity is an error, never a new Specification.
func TestDefine_ExistingSpecification_AndUnknownIdentity(t *testing.T) {
	d := newDefineEnv(t)
	if code, out := d.run("ready", "define", reserveRequest); code != 0 {
		t.Fatalf("define: %s", out)
	}

	code, out := d.run("ready", "define", "spec-001", "members may cancel a reservation")
	if code != 0 || strings.Contains(out, "Created") || strings.Contains(out, "Continuing") {
		t.Fatalf("defining an existing Specification must not create or continue another, exit %d:\n%s", code, out)
	}
	if !strings.Contains(d.lastPrompt(), "The governing Specification is: SPEC-001") ||
		!strings.Contains(d.lastPrompt(), "in their own words:\nmembers may cancel a reservation") {
		t.Fatalf("expected SPEC-001 with the change handed to the Agent:\n%s", d.lastPrompt())
	}

	code, out = d.run("ready", "define", "SPEC-009")
	if code == 0 || !strings.Contains(out, "SPEC-009 does not exist. To start a new feature, describe it instead") {
		t.Fatalf("expected an unknown identity to be refused, exit %d:\n%s", code, out)
	}
	if code, out := d.run("ready", "define", "SPEC-001", "--title", "x"); code == 0 || !strings.Contains(out, "apply only when starting a new Specification") {
		t.Fatalf("--title must be refused with an existing identity, exit %d:\n%s", code, out)
	}
	if len(d.specFiles()) != 1 {
		t.Fatalf("expected still one Specification, got %v", d.specFiles())
	}
}

// A request matching a Specification that has been approved is refused, pointing to how to change it.
func TestDefine_RequestMatchingApprovedSpecification_Refused(t *testing.T) {
	d := newDefineEnv(t)
	if code, out := d.run("ready", "define", reserveRequest); code != 0 {
		t.Fatalf("define: %s", out)
	}
	path := filepath.Join(d.dir, ".gnomon", "specifications", d.specFiles()[0])
	content, _ := os.ReadFile(path)
	filled := strings.Replace(string(content), "## Acceptance Criteria\n", "## Acceptance Criteria\n\n1. A member can reserve an item that is out.\n", 1)
	if err := os.WriteFile(path, []byte(filled), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out := runGnomonOutput(t, d.dir, d.env, "approve", "SPEC-001"); code != 0 {
		t.Fatalf("approve: %s", out)
	}

	code, out := d.run("ready", "define", reserveRequest)
	if code == 0 || !strings.Contains(out, `already exists and has been approved. To change it: gnomon define SPEC-001 "<the change>"`) {
		t.Fatalf("expected a refusal, exit %d:\n%s", code, out)
	}
	if len(d.specFiles()) != 1 {
		t.Fatalf("no duplicate of an approved Specification may be created: %v", d.specFiles())
	}
}

// A blocked Define leaves the Draft with its request and says how to continue; continuing works.
func TestDefine_Blocked_LeavesRecoverableDraft(t *testing.T) {
	d := newDefineEnv(t)
	code, out := d.run("blocked", "define", "Late fees for tools returned late")
	if code == 0 {
		t.Fatalf("a blocked Define must not exit 0:\n%s", out)
	}
	for _, want := range []string{"Created SPEC-001 — Late fees for tools returned late (Draft)",
		"SPEC-001 stays a Draft with your request saved in it. Continue it with: gnomon define SPEC-001"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in:\n%s", want, out)
		}
	}
	if code, out := runGnomonOutput(t, d.dir, d.env, "status"); code != 0 || !strings.Contains(out, "SPEC-001: Draft") {
		t.Fatalf("expected SPEC-001 listed as a Draft, exit %d:\n%s", code, out)
	}
	if code, out := d.run("ready", "define", "SPEC-001"); code != 0 || !strings.Contains(out, "Ready for approval") {
		t.Fatalf("continuing the Draft must work, exit %d:\n%s", code, out)
	}
	if len(d.specFiles()) != 1 {
		t.Fatalf("expected one Specification, got %v", d.specFiles())
	}
}

// An interrupted Define (no result) leaves the Draft, recoverable; a Draft holding only its request
// is still "nothing to approve yet".
func TestDefine_Interrupted_LeavesRecoverableDraftNotApprovable(t *testing.T) {
	d := newDefineEnv(t)
	code, out := d.run("crash", "define", reserveRequest)
	if code == 0 || !strings.Contains(out, "SPEC-001 stays a Draft with your request saved in it. Continue it with: gnomon define SPEC-001") {
		t.Fatalf("expected a recoverable Draft after an interrupted Define, exit %d:\n%s", code, out)
	}
	if code, out := runGnomonOutput(t, d.dir, d.env, "approve", "SPEC-001"); code == 0 || !strings.Contains(out, "nothing to approve yet") {
		t.Fatalf("a Draft holding only its request must not be approvable, exit %d:\n%s", code, out)
	}
	if code, out := d.run("ready", "define", reserveRequest); code != 0 || !strings.Contains(out, "Continuing SPEC-001") {
		t.Fatalf("repeating the request after an interruption continues the Draft, exit %d:\n%s", code, out)
	}
}
