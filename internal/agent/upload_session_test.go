package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
)

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func newMgr(t *testing.T, quota, maxFile int64) *UploadManager {
	t.Helper()
	m, err := NewUploadManager(t.TempDir(), quota, maxFile)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestUploadSessionResumableFinalizeAndHash(t *testing.T) {
	m := newMgr(t, 1<<20, 1<<20)
	content := []byte("abcdefghij0123456789ZZZZZ") // 25 bytes
	sum := sha256Hex(content)

	s, err := m.StartUpload("org-a", "prj", "data.bin", int64(len(content)), 10, sum)
	if err != nil || s.State != UploadReceiving {
		t.Fatalf("start: %+v err=%v", s, err)
	}
	// Append sequencial [0:10]; consulta ReceivedBytes (resume) e continua.
	part, err := m.AppendChunk("org-a", s.ID, 0, content[:10])
	if err != nil || part.ReceivedBytes != 10 {
		t.Fatalf("chunk1: %+v err=%v", part, err)
	}
	// Finalize incompleto -> erro.
	if _, err := m.FinalizeUpload("org-a", s.ID); !errors.Is(err, ErrUploadIncomplete) {
		t.Fatalf("incompleto err=%v", err)
	}
	// Retoma a partir de ReceivedBytes.
	if _, err := m.AppendChunk("org-a", s.ID, part.ReceivedBytes, content[10:]); err != nil {
		t.Fatal(err)
	}
	done, err := m.FinalizeUpload("org-a", s.ID)
	if err != nil || done.State != UploadCompleted || done.ExpectedSHA256 != sum {
		t.Fatalf("finalize: %+v err=%v", done, err)
	}
	got, err := os.ReadFile(done.FinalPath)
	if err != nil || string(got) != string(content) {
		t.Fatalf("final content mismatch: err=%v", err)
	}
}

func TestUploadSessionHashMismatch(t *testing.T) {
	m := newMgr(t, 1<<20, 1<<20)
	content := []byte("hello world")
	s, err := m.StartUpload("org-a", "p", "f.bin", int64(len(content)), 0, sha256Hex([]byte("different")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.AppendChunk("org-a", s.ID, 0, content); err != nil {
		t.Fatal(err)
	}
	if _, err := m.FinalizeUpload("org-a", s.ID); !errors.Is(err, ErrUploadHashMismatch) {
		t.Fatalf("hash mismatch err=%v", err)
	}
}

func TestUploadSessionQuotaAndSizeAndFilename(t *testing.T) {
	m := newMgr(t, 100, 1000) // quota 100 bytes
	if _, err := m.StartUpload("org-a", "p", "big.bin", 200, 0, ""); !errors.Is(err, ErrUploadQuotaExceeded) {
		t.Fatalf("quota err=%v", err)
	}
	if _, err := m.StartUpload("org-a", "p", "zero.bin", 0, 0, ""); !errors.Is(err, ErrUploadSizeInvalid) {
		t.Fatalf("size err=%v", err)
	}
	if _, err := m.StartUpload("org-a", "p", "../evil", 10, 0, ""); !errors.Is(err, ErrUploadFilename) {
		t.Fatalf("filename err=%v", err)
	}
}

func TestUploadSessionInvalidChunkAndCancel(t *testing.T) {
	m := newMgr(t, 1<<20, 1<<20)
	s, err := m.StartUpload("org-a", "p", "f.bin", 10, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	// Chunk fora dos limites.
	if _, err := m.AppendChunk("org-a", s.ID, 5, []byte("0123456789")); !errors.Is(err, ErrUploadInvalidChunk) {
		t.Fatalf("chunk oob err=%v", err)
	}
	cancelled, err := m.CancelUpload("org-a", s.ID)
	if err != nil || cancelled.State != UploadCancelled {
		t.Fatalf("cancel: %+v err=%v", cancelled, err)
	}
	// Apos cancelar, append falha (nao receiving).
	if _, err := m.AppendChunk("org-a", s.ID, 0, []byte("x")); !errors.Is(err, ErrUploadNotReceiving) {
		t.Fatalf("append apos cancel err=%v", err)
	}
}

func TestUploadSessionRejectsCrossTenant(t *testing.T) {
	m := newMgr(t, 1<<20, 1<<20)
	s, err := m.StartUpload("org-a", "p", "f.bin", 4, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.GetUploadForOrganization("org-b", s.ID); !errors.Is(err, ErrUploadForbidden) {
		t.Fatalf("get cross-tenant err=%v", err)
	}
	if _, err := m.AppendChunk("org-b", s.ID, 0, []byte("data")); !errors.Is(err, ErrUploadForbidden) {
		t.Fatalf("append cross-tenant err=%v", err)
	}
	if _, err := m.CancelUpload("org-b", s.ID); !errors.Is(err, ErrUploadForbidden) {
		t.Fatalf("cancel cross-tenant err=%v", err)
	}
}
