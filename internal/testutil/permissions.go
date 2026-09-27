// Package testutil holds small test-only helpers shared across internal packages' test files.
// Never imported by production code.
package testutil

import (
	"os"
	"runtime"
	"testing"
)

// SkipIfPermissionsNotEnforced skips t when a read-only-directory permission check would not
// actually be enforced: running as root (permission checks bypassed) or on Windows (os.Chmod on a
// directory does not prevent creating files in it there).
func SkipIfPermissionsNotEnforced(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod on a directory does not prevent writes to it on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission checks")
	}
}
