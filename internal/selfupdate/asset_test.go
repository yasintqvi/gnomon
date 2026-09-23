package selfupdate

import "testing"

func testAssets() []Asset {
	return []Asset{
		{Name: "checksums.txt", URL: "https://example.test/checksums.txt"},
		{Name: "gnomon_1.1.0_darwin_amd64.tar.gz", URL: "https://example.test/darwin_amd64.tar.gz"},
		{Name: "gnomon_1.1.0_darwin_arm64.tar.gz", URL: "https://example.test/darwin_arm64.tar.gz"},
		{Name: "gnomon_1.1.0_linux_amd64.tar.gz", URL: "https://example.test/linux_amd64.tar.gz"},
		{Name: "gnomon_1.1.0_linux_arm64.tar.gz", URL: "https://example.test/linux_arm64.tar.gz"},
		{Name: "gnomon_1.1.0_windows_amd64.zip", URL: "https://example.test/windows_amd64.zip"},
		{Name: "gnomon_1.1.0_windows_arm64.zip", URL: "https://example.test/windows_arm64.zip"},
	}
}

func TestSelectAsset_MatchesExactPlatform(t *testing.T) {
	cases := []struct{ goos, goarch, wantName string }{
		{"darwin", "amd64", "gnomon_1.1.0_darwin_amd64.tar.gz"},
		{"darwin", "arm64", "gnomon_1.1.0_darwin_arm64.tar.gz"},
		{"linux", "amd64", "gnomon_1.1.0_linux_amd64.tar.gz"},
		{"linux", "arm64", "gnomon_1.1.0_linux_arm64.tar.gz"},
		{"windows", "amd64", "gnomon_1.1.0_windows_amd64.zip"},
		{"windows", "arm64", "gnomon_1.1.0_windows_arm64.zip"},
	}
	for _, c := range cases {
		got, err := selectAsset(testAssets(), c.goos, c.goarch)
		if err != nil {
			t.Fatalf("%s/%s: unexpected error: %v", c.goos, c.goarch, err)
		}
		if got.Name != c.wantName {
			t.Errorf("%s/%s: got %q, want %q", c.goos, c.goarch, got.Name, c.wantName)
		}
	}
}

// TestSelectAsset_UnsupportedPlatform_FailsClearly proves a platform outside GoReleaser's own
// build matrix (.goreleaser.yaml's goos/goarch lists) is refused rather than guessed at.
func TestSelectAsset_UnsupportedPlatform_FailsClearly(t *testing.T) {
	_, err := selectAsset(testAssets(), "freebsd", "arm")
	if err == nil {
		t.Fatal("expected an unsupported platform to be refused")
	}
}

// TestSelectAsset_DoesNotConfuseChecksumsFile proves the checksums.txt entry, which shares no
// platform suffix with any real asset, never accidentally matches.
func TestSelectAsset_DoesNotConfuseChecksumsFile(t *testing.T) {
	assets := []Asset{{Name: "checksums.txt", URL: "https://example.test/checksums.txt"}}
	if _, err := selectAsset(assets, "linux", "amd64"); err == nil {
		t.Fatal("expected no match when only checksums.txt is present")
	}
}

func TestFindChecksumsAsset_Found(t *testing.T) {
	got, err := findChecksumsAsset(testAssets())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "checksums.txt" {
		t.Errorf("got %q, want checksums.txt", got.Name)
	}
}

func TestFindChecksumsAsset_Missing(t *testing.T) {
	_, err := findChecksumsAsset([]Asset{{Name: "gnomon_1.1.0_linux_amd64.tar.gz"}})
	if err == nil {
		t.Fatal("expected an error when no checksums.txt asset is present")
	}
}

func TestExecutableName(t *testing.T) {
	if got := executableName("windows"); got != "gnomon.exe" {
		t.Errorf("windows: got %q, want gnomon.exe", got)
	}
	for _, goos := range []string{"darwin", "linux"} {
		if got := executableName(goos); got != "gnomon" {
			t.Errorf("%s: got %q, want gnomon", goos, got)
		}
	}
}

func TestArchiveExt(t *testing.T) {
	if got := archiveExt("windows"); got != ".zip" {
		t.Errorf("windows: got %q, want .zip", got)
	}
	for _, goos := range []string{"darwin", "linux"} {
		if got := archiveExt(goos); got != ".tar.gz" {
			t.Errorf("%s: got %q, want .tar.gz", goos, got)
		}
	}
}
