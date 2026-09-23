package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

// extractExecutable returns execName's raw bytes from archiveBytes — a flat archive (no
// directory prefix; confirmed against a real published archive's contents) containing just the
// built binary and LICENSE, per .goreleaser.yaml's archives.files. archiveName's extension
// selects tar.gz vs zip handling; any other extension is refused rather than guessed at.
func extractExecutable(archiveBytes []byte, archiveName, execName string) ([]byte, error) {
	switch {
	case strings.HasSuffix(archiveName, ".zip"):
		return extractFromZip(archiveBytes, execName)
	case strings.HasSuffix(archiveName, ".tar.gz"):
		return extractFromTarGz(archiveBytes, execName)
	default:
		return nil, fmt.Errorf("unrecognized archive format: %s", archiveName)
	}
}

func extractFromZip(data []byte, execName string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading archive: %w", err)
	}
	for _, f := range r.File {
		if f.Name != execName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("reading %s from archive: %w", execName, err)
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("archive does not contain %s", execName)
}

func extractFromTarGz(data []byte, execName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("reading archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading archive: %w", err)
		}
		if hdr.Name != execName {
			continue
		}
		return io.ReadAll(tr)
	}
	return nil, fmt.Errorf("archive does not contain %s", execName)
}
