package result

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gnomon/internal/contract"
)

func testResultContract() contract.ResultContract {
	return contract.ResultContract{
		TerminalPath: "outcome",
		Schema: map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"outcome"},
			"properties": map[string]interface{}{
				"outcome": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"IMPLEMENTATION_COMPLETE", "BLOCKED"},
				},
			},
		},
	}
}

func writeEnvelope(t *testing.T, path, workflow, runID string, payload map[string]interface{}) {
	t.Helper()
	payloadBytes, _ := json.Marshal(payload)
	env := map[string]interface{}{
		"workflow": workflow,
		"run_id":   runID,
		"payload":  json.RawMessage(payloadBytes),
	}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTransientDir_CreatesSelfContainedGitignore(t *testing.T) {
	root := t.TempDir()
	dir, err := TransientDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) != transientDirName {
		t.Fatalf("unexpected transient dir name: %s", dir)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "*\n!.gitignore\n" {
		t.Fatalf("unexpected gitignore content: %q", data)
	}
}

func TestConsume_ValidResult_ExtractsTerminalAndDeletesFile(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, err := Destination(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	writeEnvelope(t, dest, "implementation", runID, map[string]interface{}{
		"outcome":   "IMPLEMENTATION_COMPLETE",
		"delivered": "a file",
	})

	outcome, err := Consume(dest, "implementation", runID, testResultContract())
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Terminal != "IMPLEMENTATION_COMPLETE" {
		t.Fatalf("unexpected terminal value: %s", outcome.Terminal)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected the transient result file to be deleted after successful consumption")
	}
}

func TestConsume_WorkflowIdentityMismatch(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, _ := Destination(root, runID)
	writeEnvelope(t, dest, "verification", runID, map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"})

	if _, err := Consume(dest, "implementation", runID, testResultContract()); err == nil {
		t.Fatalf("expected a workflow identity mismatch to be rejected")
	}
}

func TestConsume_RunIDMismatch(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, _ := Destination(root, runID)
	writeEnvelope(t, dest, "implementation", "some-other-run-id", map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"})

	if _, err := Consume(dest, "implementation", runID, testResultContract()); err == nil {
		t.Fatalf("expected a run_id mismatch to be rejected (stale/misrouted result)")
	}
}

func TestConsume_UnknownOutcomeValue_RejectedBySchema(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, _ := Destination(root, runID)
	writeEnvelope(t, dest, "implementation", runID, map[string]interface{}{"outcome": "BANANA"})

	if _, err := Consume(dest, "implementation", runID, testResultContract()); err == nil {
		t.Fatalf("expected an unknown Outcome value to fail Result Contract validation")
	}
}

func TestConsume_MissingResult_ProtocolFailure(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, _ := Destination(root, runID)
	if _, err := Consume(dest, "implementation", runID, testResultContract()); err == nil {
		t.Fatalf("expected a missing result file to be reported as a protocol failure")
	}
}

func TestConsume_MalformedJSON_LeftInPlaceForDiagnostics(t *testing.T) {
	root := t.TempDir()
	runID, _ := NewRunID()
	dest, _ := Destination(root, runID)
	if err := os.WriteFile(dest, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Consume(dest, "implementation", runID, testResultContract()); err == nil {
		t.Fatalf("expected malformed JSON to be rejected")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected the malformed file to remain for diagnostics, not be deleted")
	}
}
