// Package selfupdate implements `gnomon update`: downloading and installing the latest stable
// Gnomon release over the currently running executable. It is deliberately independent of every
// project-engineering concern — no .gnomon, no Specifications, no Agents, no workflow execution,
// no Result Protocol — so it works identically from any directory, with or without a Gnomon
// project present, exactly like `gnomon agent`.
package selfupdate

import (
	"fmt"
	"io"

	"golang.org/x/mod/semver"

	"gnomon/internal/present"
)

// Options carries everything Run needs from its caller. Every field the underlying logic must be
// deterministic and offline-testable over is supplied here rather than discovered internally
// (runtime.GOOS/GOARCH, os.Executable) — so cmd/gnomon resolves real environment facts exactly
// once, at the CLI edge, matching how it already resolves the working directory before calling
// into internal/orchestrate.
type Options struct {
	// CurrentVersion is the running binary's own version (main.version, "dev" for a local build).
	CurrentVersion string
	// GOOS/GOARCH select the release artifact; callers pass runtime.GOOS/runtime.GOARCH.
	GOOS, GOARCH string
	// ExecutablePath is the file to replace — see ExecutablePath's own doc for how callers should
	// resolve it.
	ExecutablePath string
	// Fetcher is the network boundary; NewGitHubFetcher() outside tests.
	Fetcher Fetcher
}

// Run performs the entire update: check the latest stable release, compare it against
// CurrentVersion, and — only if it is actually newer — download, verify, extract, and install it.
// Progress is written to out as each step completes; out sees nothing at all when already
// up to date. The installed executable is never touched until the downloaded artifact has already
// passed checksum verification.
func Run(out io.Writer, opts Options) (*present.Report, error) {
	rel, err := opts.Fetcher.LatestRelease(owner, repo)
	if err != nil {
		return nil, err
	}
	if rel.Draft || rel.Prerelease {
		return nil, fmt.Errorf("the latest GitHub release (%s) is a draft or prerelease; no stable release is available", rel.TagName)
	}

	if !isNewer(opts.CurrentVersion, rel.TagName) {
		return &present.Report{
			Outcome: present.Success,
			Summary: fmt.Sprintf("Gnomon %s is already up to date.", opts.CurrentVersion),
		}, nil
	}

	asset, err := selectAsset(rel.Assets, opts.GOOS, opts.GOARCH)
	if err != nil {
		return nil, err
	}
	checksumsAsset, err := findChecksumsAsset(rel.Assets)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(out, "Updating Gnomon...")
	fmt.Fprintln(out)

	archiveBytes, err := opts.Fetcher.Download(asset.URL)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", asset.Name, err)
	}
	fmt.Fprintln(out, "✓ Downloaded release")

	checksumBytes, err := opts.Fetcher.Download(checksumsAsset.URL)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", checksumsAssetName, err)
	}
	sums, err := parseChecksums(checksumBytes)
	if err != nil {
		return nil, err
	}
	expected, ok := sums[asset.Name]
	if !ok {
		return nil, fmt.Errorf("%s has no checksum entry for %s", checksumsAssetName, asset.Name)
	}
	if err := verifyChecksum(archiveBytes, expected); err != nil {
		return nil, err
	}
	fmt.Fprintln(out, "✓ Verified checksum")

	execBytes, err := extractExecutable(archiveBytes, asset.Name, executableName(opts.GOOS))
	if err != nil {
		return nil, fmt.Errorf("extracting %s: %w", asset.Name, err)
	}

	if err := replaceExecutable(opts.ExecutablePath, execBytes); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "✓ Installed Gnomon %s\n", rel.TagName)
	fmt.Fprintln(out)

	return &present.Report{
		Outcome: present.Success,
		Summary: "Gnomon updated successfully",
		Sections: []present.Section{
			{Label: "Version", Body: fmt.Sprintf("%s → %s", opts.CurrentVersion, rel.TagName)},
		},
	}, nil
}

// isNewer reports whether latest is a real upgrade over current, using proper semantic-version
// comparison (golang.org/x/mod/semver) rather than string equality. current is treated as
// out of date whenever it is not itself a valid vMAJOR.MINOR.PATCH version — the case for an
// ordinary local "dev" build — so `gnomon update` always offers to move such a build onto a real
// release rather than refusing to compare it.
func isNewer(current, latest string) bool {
	if !semver.IsValid(current) {
		return true
	}
	return semver.Compare(current, latest) < 0
}
