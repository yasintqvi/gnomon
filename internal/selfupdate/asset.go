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

// executableName returns the filename Gnomon's own executable has inside a release archive for
// goos — confirmed by inspecting a real published archive's contents, not assumed: GoReleaser
// names the built binary "gnomon" (builds.binary in .goreleaser.yaml) and appends ".exe" for
// windows automatically.
func executableName(goos string) string {
	if goos == "windows" {
		return "gnomon.exe"
	}
	return "gnomon"
}

// selectAsset finds the one release Asset matching goos/goarch, from GoReleaser's own
// archives.name_template: "{{.ProjectName}}_{{.Version}}_{{.Os}}_{{.Arch}}" — i.e.
// "gnomon_<version>_<goos>_<goarch>.<ext>", where .Os/.Arch are Go's own runtime.GOOS/GOARCH
// values verbatim (goreleaser applies no name mangling here), also confirmed against real
// published asset names. Matching by suffix (rather than reconstructing the full name, which
// would require knowing the release's de-"v"-prefixed version string) keeps this independent of
// exactly how the version segment is spelled.
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
