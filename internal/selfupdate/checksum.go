package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// parseChecksums reads GoReleaser's own checksums.txt format (checksum.algorithm: sha256 in
// .goreleaser.yaml) — "<hex digest><spaces><filename>" per line, confirmed against a real
// published checksums.txt — into a filename → hex-digest map.
func parseChecksums(data []byte) (map[string]string, error) {
	sums := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed checksums entry: %q", line)
		}
		sums[fields[1]] = strings.ToLower(fields[0])
	}
	return sums, nil
}

// verifyChecksum reports a non-nil error unless data's own SHA-256 digest matches expectedHex —
// the mandatory gate between a download and ever installing it.
func verifyChecksum(data []byte, expectedHex string) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, expectedHex) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, got)
	}
	return nil
}
