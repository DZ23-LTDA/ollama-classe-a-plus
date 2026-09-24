package agent

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type zipEntry struct {
	name    string
	content string
	symlink bool
}

func writeTestZip(t *testing.T, entries []zipEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.symlink {
			hdr.SetMode(os.ModeSymlink | 0o777)
		} else {
			hdr.SetMode(0o644)
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSafeExtractZipHappyPath(t *testing.T) {
	zp := writeTestZip(t, []zipEntry{
		{name: "index.html", content: "<h1>ok</h1>"},
		{name: "assets/app.js", content: "console.log(1)"},
	})
	dest := filepath.Join(t.TempDir(), "out")
	rep, err := SafeExtractZip(zp, dest, DefaultExtractLimits())
	if err != nil {
		t.Fatalf("extract err=%v", err)
	}
	if rep.Files != 2 {
		t.Fatalf("files=%d report=%+v", rep.Files, rep)
	}
	if _, err := os.Stat(filepath.Join(dest, "assets", "app.js")); err != nil {
		t.Fatalf("expected extracted file: %v", err)
	}
}

func TestSafeExtractZipRejectsTraversalAndAbsolute(t *testing.T) {
	base := t.TempDir()
	dest := filepath.Join(base, "out")
	for _, bad := range []string{"../pwned", "a/../../pwned", "/etc/pwned", "..\\pwned"} {
		zp := writeTestZip(t, []zipEntry{{name: bad, content: "x"}, {name: "ok.txt", content: "y"}})
		if _, err := SafeExtractZip(zp, dest, DefaultExtractLimits()); !errors.Is(err, ErrArchivePathEscapes) {
			t.Fatalf("%q: esperava zip-slip, got %v", bad, err)
		}
	}
	// Nada pode ter escapado para o diretorio-pai.
	if _, err := os.Stat(filepath.Join(base, "pwned")); err == nil {
		t.Fatal("arquivo escapou para o diretorio-pai")
	}
}

func TestSafeExtractZipRejectsSymlink(t *testing.T) {
	zp := writeTestZip(t, []zipEntry{{name: "link", content: "/etc/passwd", symlink: true}})
	dest := filepath.Join(t.TempDir(), "out")
	if _, err := SafeExtractZip(zp, dest, DefaultExtractLimits()); !errors.Is(err, ErrArchiveSymlink) {
		t.Fatalf("esperava symlink rejeitado, got %v", err)
	}
}

func TestSafeExtractZipEnforcesFileCount(t *testing.T) {
	zp := writeTestZip(t, []zipEntry{{name: "a", content: "1"}, {name: "b", content: "2"}, {name: "c", content: "3"}})
	dest := filepath.Join(t.TempDir(), "out")
	lim := DefaultExtractLimits()
	lim.MaxFiles = 2
	if _, err := SafeExtractZip(zp, dest, lim); !errors.Is(err, ErrArchiveTooManyFiles) {
		t.Fatalf("esperava too-many-files, got %v", err)
	}
}

func TestSafeExtractZipEnforcesPerFileSize(t *testing.T) {
	zp := writeTestZip(t, []zipEntry{{name: "big.txt", content: strings.Repeat("ab", 5000)}}) // ~10 KB, pouco compressivel
	dest := filepath.Join(t.TempDir(), "out")
	lim := DefaultExtractLimits()
	lim.MaxFileBytes = 1024
	lim.MaxRatio = 1 << 30 // isola o teste de tamanho do de ratio
	if _, err := SafeExtractZip(zp, dest, lim); !errors.Is(err, ErrArchiveFileTooBig) {
		t.Fatalf("esperava file-too-big, got %v", err)
	}
}

func TestSafeExtractZipEnforcesRatio(t *testing.T) {
	// 200 KB de zeros comprime muito -> ratio alto (zip bomb).
	zp := writeTestZip(t, []zipEntry{{name: "bomb.bin", content: strings.Repeat("\x00", 200*1024)}})
	dest := filepath.Join(t.TempDir(), "out")
	lim := DefaultExtractLimits()
	lim.MaxRatio = 10
	lim.MaxFileBytes = 1 << 30 // isola ratio do tamanho
	if _, err := SafeExtractZip(zp, dest, lim); !errors.Is(err, ErrArchiveRatio) {
		t.Fatalf("esperava ratio/zip-bomb, got %v", err)
	}
}
