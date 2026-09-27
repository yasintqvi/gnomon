package selfupdate

import (
	"fmt"
	"strings"
)

// checksumsAssetName is the exact, fixed name GoReleaser's own checksum.name_template
// (.goreleaser.yaml) produces — verified against a real published release, not assumed.
const checksumsAssetName = "checksums.txt"

// archiveExt returns the archive extension GoReleaser actually produces for goos, per
// .goreleaser.yaml's archives.formats/format_overrides ([tar.gz] for every platform, overridden
// to [zip] for windows).
func archiveExt(goos string) string {
	if goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// executableName returns the filename Gnomon's executable has inside a release archive for goos —
// confirmed against a real published archive, not assumed.
func executableName(goos string) string {
	if goos == "windows" {
		return "gnomon.exe"
	}
	return "gnomon"
}

// selectAsset finds the release Asset matching goos/goarch: "gnomon_<version>_<goos>_<goarch>.<ext>"
// (confirmed against real published asset names). Matched by suffix rather than reconstructing the
// full name, so this stays independent of exactly how the version segment is spelled.
func selectAsset(assets []Asset, goos, goarch string) (Asset, error) {
	suffix := fmt.Sprintf("_%s_%s%s", goos, goarch, archiveExt(goos))
	for _, a := range assets {
		if strings.HasPrefix(a.Name, "gnomon_") && strings.HasSuffix(a.Name, suffix) {
			return a, nil
		}
	}
	return Asset{}, fmt.Errorf("no release artifact available for %s/%s", goos, goarch)
}

// findChecksumsAsset locates the release's own checksums.txt among its assets.
func findChecksumsAsset(assets []Asset) (Asset, error) {
	for _, a := range assets {
		if a.Name == checksumsAssetName {
			return a, nil
		}
	}
	return Asset{}, fmt.Errorf("release is missing %s", checksumsAssetName)
}
