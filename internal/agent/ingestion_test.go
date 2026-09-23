package agent

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentIngestorChunksTextAndPersistsProvenance(t *testing.T) {
	root := t.TempDir()
	contextStore, err := NewContextStore(filepath.Join(root, "context"))
	if err != nil {
		t.Fatal(err)
	}
	project, err := contextStore.CreateProject("Docs", root)
	if err != nil {
		t.Fatal(err)
	}
	content := strings.Repeat("architecture evidence ", 200)
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	memories, err := (DocumentIngestor{Context: contextStore}).Ingest(context.Background(), DocumentIngestRequest{ProjectID: project.ID, Paths: []string{"notes.md"}, ChunkSize: 200, ChunkOverlap: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) < 2 || !strings.HasPrefix(memories[0].Source, "notes.md#chunk-") {
		t.Fatalf("memories=%+v", memories)
	}
	if _, err := (DocumentIngestor{Context: contextStore}).Ingest(context.Background(), DocumentIngestRequest{ProjectID: project.ID, Paths: []string{"../outside.txt"}}); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestDocumentIngestorRejectsWorkspaceOutsideProjectAndSymlink(t *testing.T) {
	root := t.TempDir()
	contextStore, err := NewContextStore("")
	if err != nil {
		t.Fatal(err)
	}
	project, err := contextStore.CreateProject("Docs", root)
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	ingestor := DocumentIngestor{Context: contextStore}
	if _, err := ingestor.Ingest(context.Background(), DocumentIngestRequest{ProjectID: project.ID, Workspace: outside, Paths: []string{"secret.txt"}}); err == nil || !strings.Contains(err.Error(), "inside the project root") {
		t.Fatalf("outside workspace error = %v", err)
	}
	linked := filepath.Join(root, "linked")
	if err := os.Symlink(outside, linked); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := ingestor.Ingest(context.Background(), DocumentIngestRequest{ProjectID: project.ID, Workspace: linked, Paths: []string{"secret.txt"}}); err == nil || !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("symlink workspace error = %v", err)
	}
}

func TestDocumentIngestorRejectsPersistedProjectRootOutsideRuntimeWorkspace(t *testing.T) {
	runtimeWorkspace := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("outside runtime workspace"), 0o600); err != nil {
		t.Fatal(err)
	}
	storeRoot := t.TempDir()
	legacyStore, err := NewContextStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}
	project, err := legacyStore.CreateProject("Legacy", outside)
	if err != nil {
		t.Fatal(err)
	}
	contextStore, err := NewContextStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}
	_, err = (DocumentIngestor{Context: contextStore, WorkspaceRoot: runtimeWorkspace}).Ingest(context.Background(), DocumentIngestRequest{ProjectID: project.ID, Paths: []string{"secret.txt"}})
	if err == nil || !strings.Contains(err.Error(), "outside the runtime workspace") {
		t.Fatalf("outside project root was accepted: %v", err)
	}
}

func TestDocumentIngestorHonorsCancellationBeforeReading(t *testing.T) {
	root := t.TempDir()
	contextStore, err := NewContextStore("")
	if err != nil {
		t.Fatal(err)
	}
	project, err := contextStore.CreateProject("Docs", root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (DocumentIngestor{Context: contextStore}).Ingest(ctx, DocumentIngestRequest{ProjectID: project.ID, Paths: []string{"notes.md"}}); err == nil || err != context.Canceled {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestReadDOCXRejectsCompressedEntryOverBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.docx")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("<document>content larger than the limit</document>")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readDOCX(context.Background(), path, 8); err == nil || !strings.Contains(err.Error(), "exceeds byte budget") {
		t.Fatalf("DOCX budget error = %v", err)
	}
}
