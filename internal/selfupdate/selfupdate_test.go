package selfupdate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gnomon/internal/present"
)

// fakeFetcher is the test double for Fetcher — the one seam Run depends on — so every test here
// runs fully offline and deterministically, never reaching a live GitHub endpoint.
type fakeFetcher struct {
	release    Release
	releaseErr error

	downloads    map[string][]byte
	downloadErrs map[string]error
}

func (f *fakeFetcher) LatestRelease(owner, repo string) (Release, error) {
	return f.release, f.releaseErr
}

func (f *fakeFetcher) Download(url string) ([]byte, error) {
	if err, ok := f.downloadErrs[url]; ok {
		return nil, err
	}
	if b, ok := f.downloads[url]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("fakeFetcher: no data configured for %s", url)
}

// --- isNewer / semantic version comparison ---

func TestIsNewer_ProperSemVerComparison(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.0.0", "v1.1.0", true},
		{"v1.1.0", "v1.1.0", false},
		{"v1.2.0", "v1.1.0", false},
		{"v1.9.0", "v1.10.0", true}, // numeric, not lexicographic, comparison
		{"v1.0.9", "v1.0.10", true}, // same
		{"v2.0.0", "v1.99.99", false},
	}
	for _, c := range cases {
		if got := isNewer(c.current, c.latest); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func TestIsNewer_InvalidCurrentVersion_TreatedAsOutOfDate(t *testing.T) {
	if !isNewer("dev", "v1.1.0") {
		t.Fatal("expected a non-semver current version (e.g. a local \"dev\" build) to be treated as needing an update")
	}
}

// --- Run: already up to date ---

func TestRun_AlreadyUpToDate_NoDownloadNoInstall(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	if err := os.WriteFile(target, []byte("current binary"), 0o755); err != nil {
		t.Fatalf("seeding executable: %v", err)
	}

	fetcher := &fakeFetcher{release: Release{TagName: "v1.1.0"}}
	var out bytes.Buffer
	rep, err := Run(&out, Options{
		CurrentVersion: "v1.1.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
	if rep.Summary != "Gnomon v1.1.0 is already up to date." {
		t.Fatalf("unexpected summary: %q", rep.Summary)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no progress output when already up to date, got %q", out.String())
	}
	got, _ := os.ReadFile(target)
	if string(got) != "current binary" {
		t.Fatalf("expected the executable untouched, got %q", got)
	}
}

// --- Run: newer stable version available (full happy path) ---

func TestRun_NewerVersionAvailable_DownloadsVerifiesAndInstalls(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("seeding executable: %v", err)
	}

	archive := makeTarGz(t, map[string][]byte{
		"LICENSE": []byte("MIT"),
		"gnomon":  []byte("new binary content"),
	})
	checksums := []byte(sha256HexForTest(archive) + "  gnomon_1.1.0_linux_amd64.tar.gz\n")

	fetcher := &fakeFetcher{
		release: Release{
			TagName: "v1.1.0",
			Assets: []Asset{
				{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
				{Name: "checksums.txt", URL: "https://example.test/checksums"},
			},
		},
		downloads: map[string][]byte{
			"https://example.test/archive":   archive,
			"https://example.test/checksums": checksums,
		},
	}

	var out bytes.Buffer
	rep, err := Run(&out, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Summary != "Gnomon updated successfully" {
		t.Fatalf("unexpected summary: %q", rep.Summary)
	}
	if len(rep.Sections) != 1 || rep.Sections[0].Body != "v1.0.0 → v1.1.0" {
		t.Fatalf("expected a Version section describing the transition, got %+v", rep.Sections)
	}

	progress := out.String()
	for _, want := range []string{"Updating Gnomon...", "✓ Downloaded release", "✓ Verified checksum", "✓ Installed Gnomon v1.1.0"} {
		if !bytes.Contains([]byte(progress), []byte(want)) {
			t.Errorf("expected progress output to contain %q, got:\n%s", want, progress)
		}
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading installed executable: %v", err)
	}
	if string(got) != "new binary content" {
		t.Fatalf("expected the executable replaced with the new content, got %q", got)
	}
}

// --- Run: prerelease/draft exclusion ---

func TestRun_LatestIsPrerelease_Refused(t *testing.T) {
	fetcher := &fakeFetcher{release: Release{TagName: "v1.2.0-rc1", Prerelease: true}}
	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.1.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: "/does/not/matter", Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a prerelease latest release to be refused rather than installed")
	}
}

func TestRun_LatestIsDraft_Refused(t *testing.T) {
	fetcher := &fakeFetcher{release: Release{TagName: "v1.2.0", Draft: true}}
	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.1.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: "/does/not/matter", Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a draft latest release to be refused rather than installed")
	}
}

// --- Run: API/download failure ---

func TestRun_LatestReleaseAPIFailure(t *testing.T) {
	fetcher := &fakeFetcher{releaseErr: fmt.Errorf("network unreachable")}
	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: "/does/not/matter", Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a release-lookup failure to be surfaced")
	}
}

func TestRun_DownloadFailure_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)

	fetcher := &fakeFetcher{
		release: Release{
			TagName: "v1.1.0",
			Assets: []Asset{
				{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
				{Name: "checksums.txt", URL: "https://example.test/checksums"},
			},
		},
		downloadErrs: map[string]error{"https://example.test/archive": fmt.Errorf("connection reset")},
	}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a download failure to be surfaced")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched after a download failure, got %q", got)
	}
}

// --- Run: unsupported platform ---

func TestRun_UnsupportedPlatform_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)

	fetcher := &fakeFetcher{release: Release{TagName: "v1.1.0", Assets: []Asset{
		{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
	}}}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "plan9", GOARCH: "386",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected an unsupported platform to be refused")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched, got %q", got)
	}
}

// --- Run: checksum mismatch ---

func TestRun_ChecksumMismatch_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)

	archive := makeTarGz(t, map[string][]byte{"gnomon": []byte("new binary")})
	wrongChecksums := []byte(sha256HexForTest([]byte("not the archive")) + "  gnomon_1.1.0_linux_amd64.tar.gz\n")

	fetcher := &fakeFetcher{
		release: Release{TagName: "v1.1.0", Assets: []Asset{
			{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
			{Name: "checksums.txt", URL: "https://example.test/checksums"},
		}},
		downloads: map[string][]byte{
			"https://example.test/archive":   archive,
			"https://example.test/checksums": wrongChecksums,
		},
	}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a checksum mismatch to be refused")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched after a checksum mismatch, got %q", got)
	}
}

// --- Run: extraction failure ---

func TestRun_ExtractionFailure_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)

	archive := makeTarGz(t, map[string][]byte{"LICENSE": []byte("MIT")}) // no "gnomon" entry
	checksums := []byte(sha256HexForTest(archive) + "  gnomon_1.1.0_linux_amd64.tar.gz\n")

	fetcher := &fakeFetcher{
		release: Release{TagName: "v1.1.0", Assets: []Asset{
			{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
			{Name: "checksums.txt", URL: "https://example.test/checksums"},
		}},
		downloads: map[string][]byte{
			"https://example.test/archive":   archive,
			"https://example.test/checksums": checksums,
		},
	}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected extraction failure (no matching executable entry) to be refused")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched after an extraction failure, got %q", got)
	}
}

// --- Run: replacement failure ---

func TestRun_ReplacementFailure_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("making dir read-only: %v", err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission checks")
	}

	archive := makeTarGz(t, map[string][]byte{"gnomon": []byte("new binary")})
	checksums := []byte(sha256HexForTest(archive) + "  gnomon_1.1.0_linux_amd64.tar.gz\n")

	fetcher := &fakeFetcher{
		release: Release{TagName: "v1.1.0", Assets: []Asset{
			{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
			{Name: "checksums.txt", URL: "https://example.test/checksums"},
		}},
		downloads: map[string][]byte{
			"https://example.test/archive":   archive,
			"https://example.test/checksums": checksums,
		},
	}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected replacement to fail against an unwritable directory")
	}

	os.Chmod(dir, 0o755)
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("reading target after failed update: %v", readErr)
	}
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched after a replacement failure, got %q", got)
	}
}

// --- Run: missing checksum entry for the selected asset ---

func TestRun_NoChecksumEntryForAsset_ExecutableUntouched(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	os.WriteFile(target, []byte("old binary"), 0o755)

	archive := makeTarGz(t, map[string][]byte{"gnomon": []byte("new binary")})
	checksums := []byte(sha256HexForTest(archive) + "  some_other_file.tar.gz\n")

	fetcher := &fakeFetcher{
		release: Release{TagName: "v1.1.0", Assets: []Asset{
			{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/archive"},
			{Name: "checksums.txt", URL: "https://example.test/checksums"},
		}},
		downloads: map[string][]byte{
			"https://example.test/archive":   archive,
			"https://example.test/checksums": checksums,
		},
	}

	_, err := Run(&bytes.Buffer{}, Options{
		CurrentVersion: "v1.0.0", GOOS: "linux", GOARCH: "amd64",
		ExecutablePath: target, Fetcher: fetcher,
	})
	if err == nil {
		t.Fatal("expected a missing checksum entry to be refused")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "old binary" {
		t.Fatalf("expected the executable untouched, got %q", got)
	}
}
