// Package result implements CLI Step 5 — Result Protocol: an invocation-specific transient JSON
// envelope, consumed once after Terminal Handoff ends, then deleted. It never parses terminal
// output and never treats a workflow result as persisted state.
package result

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gnomon/internal/contract"
)

// transientDirName is the project-local, gitignored directory sibling to .gnomon/ — never nested
// inside it, per cli/RESULT_PROTOCOL.md.
const transientDirName = ".gnomon-runtime"

// NewRunID generates a collision-resistant per-invocation identifier — a runtime transport
// concern only, never Core semantics, never persisted beyond the run.
func NewRunID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// TransientDir ensures the sibling transient directory exists, with its own self-contained
// gitignore rule (touching no other .gitignore anywhere in the project), and returns its path.
func TransientDir(root string) (string, error) {
	dir := filepath.Join(root, transientDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	gitignore := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		if err := os.WriteFile(gitignore, []byte("*\n!.gitignore\n"), 0o644); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// Destination returns the path a given run's result must be published to.
func Destination(root, runID string) (string, error) {
	dir, err := TransientDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, runID+".result.json"), nil
}

type envelope struct {
	Workflow string          `json:"workflow"`
	RunID    string          `json:"run_id"`
	Payload  json.RawMessage `json:"payload"`
}

// Outcome is the fully validated, extracted result of one workflow run.
type Outcome struct {
	Terminal string
	Payload  map[string]interface{}
}

// Consume reads, structurally validates, and extracts the terminal value from the transient
// result at path, then deletes it. Any failure is returned as a distinct protocol-failure error,
// never mapped onto a Core Outcome value. On failure the file is left in place for diagnostics.
func Consume(path, expectedWorkflow, expectedRunID string, rc contract.ResultContract) (Outcome, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Outcome{}, fmt.Errorf("protocol failure: no result was produced")
		}
		return Outcome{}, fmt.Errorf("protocol failure: %w", err)
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Outcome{}, fmt.Errorf("protocol failure: malformed result: %w", err)
	}
	if env.Workflow != expectedWorkflow {
		return Outcome{}, fmt.Errorf("protocol failure: workflow identity mismatch (got %q, expected %q)", env.Workflow, expectedWorkflow)
	}
	if env.RunID != expectedRunID {
		return Outcome{}, fmt.Errorf("protocol failure: run identity mismatch (stale or misrouted result)")
	}

	var payload interface{}
	dec := json.NewDecoder(bytes.NewReader(env.Payload))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return Outcome{}, fmt.Errorf("protocol failure: malformed payload: %w", err)
	}
	if err := rc.ValidatePayload(payload); err != nil {
		return Outcome{}, fmt.Errorf("protocol failure: result contract violation: %w", err)
	}
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return Outcome{}, fmt.Errorf("protocol failure: payload is not an object")
	}
	terminal, err := rc.TerminalValue(payloadMap)
	if err != nil {
		return Outcome{}, fmt.Errorf("protocol failure: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return Outcome{}, fmt.Errorf("consumed result but failed to remove transient file: %w", err)
	}

	return Outcome{Terminal: terminal, Payload: payloadMap}, nil
}
