// Package selfupdate implements `gnomon update`: downloading and installing the latest stable
// Gnomon release over the currently running executable. Independent of every project-engineering
// concern, so it works identically from any directory, with or without a Gnomon project present.
package selfupdate

import (
	"fmt"
	"io"

	"golang.org/x/mod/semver"

	"gnomon/internal/present"
)

// Options carries everything Run needs from its caller — supplied here rather than discovered
// internally (runtime.GOOS/GOARCH, os.Executable) so the underlying logic stays deterministic and
// offline-testable.
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

// Run checks the latest stable release against CurrentVersion, and — only if newer — downloads,
// verifies, extracts, and installs it. The installed executable is never touched until the
// downloaded artifact has passed checksum verification.
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

// isNewer reports whether latest is a real upgrade over current, via semantic-version comparison.
// current is treated as out of date whenever it isn't a valid vMAJOR.MINOR.PATCH version (an
// ordinary local "dev" build), so `gnomon update` always offers to move it onto a real release.
func isNewer(current, latest string) bool {
	if !semver.IsValid(current) {
		return true
	}
	return semver.Compare(current, latest) < 0
}
