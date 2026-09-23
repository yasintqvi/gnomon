package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func sha256HexForTest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestParseChecksums_ValidFile(t *testing.T) {
	data := []byte("212e2a7b6eaa5e2f3620873fc8fd84fc7cbc65ab6ddfbadc2bb26570ebef0c2d  gnomon_1.1.0_darwin_amd64.tar.gz\n" +
		"88ff36d36a6e0938a50597b3a014bd284c0f4d038b0007d5860667298182ff29  gnomon_1.1.0_linux_arm64.tar.gz\n")
	got, err := parseChecksums(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"gnomon_1.1.0_darwin_amd64.tar.gz": "212e2a7b6eaa5e2f3620873fc8fd84fc7cbc65ab6ddfbadc2bb26570ebef0c2d",
		"gnomon_1.1.0_linux_arm64.tar.gz":  "88ff36d36a6e0938a50597b3a014bd284c0f4d038b0007d5860667298182ff29",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(want), got)
	}
	for name, sum := range want {
		if got[name] != sum {
			t.Errorf("%s: got %q, want %q", name, got[name], sum)
		}
	}
}

func TestParseChecksums_MalformedLine(t *testing.T) {
	if _, err := parseChecksums([]byte("not-a-valid-checksum-line")); err == nil {
		t.Fatal("expected a malformed line to be rejected")
	}
}

func TestParseChecksums_EmptyInput(t *testing.T) {
	got, err := parseChecksums([]byte("\n\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no entries, got %+v", got)
	}
}

func TestVerifyChecksum_Match(t *testing.T) {
	data := []byte("hello gnomon")
	sum := sha256HexForTest(data)
	if err := verifyChecksum(data, sum); err != nil {
		t.Fatalf("expected the correct checksum to verify, got: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	data := []byte("hello gnomon")
	wrong := sha256HexForTest([]byte("something else"))
	if err := verifyChecksum(data, wrong); err == nil {
		t.Fatal("expected a checksum mismatch to be rejected")
	}
}

func TestVerifyChecksum_CaseInsensitive(t *testing.T) {
	data := []byte("hello gnomon")
	sum := strings.ToUpper(sha256HexForTest(data))
	if err := verifyChecksum(data, sum); err != nil {
		t.Fatalf("expected an uppercase-hex checksum to still verify, got: %v", err)
	}
}
