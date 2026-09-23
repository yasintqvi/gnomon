package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"
)

func makeTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}); err != nil {
			t.Fatalf("writing tar header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("writing tar content: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

func makeZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("creating zip entry: %v", err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("writing zip content: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestExtractExecutable_TarGz(t *testing.T) {
	archive := makeTarGz(t, map[string][]byte{
		"LICENSE": []byte("MIT"),
		"gnomon":  []byte("fake-binary-content"),
	})
	got, err := extractExecutable(archive, "gnomon_1.1.0_linux_amd64.tar.gz", "gnomon")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-binary-content" {
		t.Fatalf("got %q, want %q", got, "fake-binary-content")
	}
}

func TestExtractExecutable_Zip(t *testing.T) {
	archive := makeZip(t, map[string][]byte{
		"LICENSE":    []byte("MIT"),
		"gnomon.exe": []byte("fake-windows-binary"),
	})
	got, err := extractExecutable(archive, "gnomon_1.1.0_windows_amd64.zip", "gnomon.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-windows-binary" {
		t.Fatalf("got %q, want %q", got, "fake-windows-binary")
	}
}

// TestExtractExecutable_MissingEntry_FailsClearly proves an archive that does not actually
// contain the expected executable name is refused, not silently treated as if it were empty.
func TestExtractExecutable_MissingEntry_FailsClearly(t *testing.T) {
	archive := makeTarGz(t, map[string][]byte{"LICENSE": []byte("MIT")})
	if _, err := extractExecutable(archive, "gnomon_1.1.0_linux_amd64.tar.gz", "gnomon"); err == nil {
		t.Fatal("expected an error when the archive has no matching executable entry")
	}
}

func TestExtractExecutable_CorruptArchive_FailsClearly(t *testing.T) {
	if _, err := extractExecutable([]byte("not an archive"), "gnomon_1.1.0_linux_amd64.tar.gz", "gnomon"); err == nil {
		t.Fatal("expected a corrupt tar.gz to be refused")
	}
	if _, err := extractExecutable([]byte("not an archive"), "gnomon_1.1.0_windows_amd64.zip", "gnomon.exe"); err == nil {
		t.Fatal("expected a corrupt zip to be refused")
	}
}

func TestExtractExecutable_UnrecognizedFormat(t *testing.T) {
	if _, err := extractExecutable([]byte("anything"), "gnomon_1.1.0_linux_amd64.tar.xz", "gnomon"); err == nil {
		t.Fatal("expected an unrecognized archive extension to be refused")
	}
}
