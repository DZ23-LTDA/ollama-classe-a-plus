package agent

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilderCreatesPreviewsExportsAndPublishes(t *testing.T) {
	service, err := NewBuilderService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	project, err := service.Create(context.Background(), BuilderSpec{Name: "Demo Game", Kind: BuilderGame, Files: map[string]string{"index.html": "<main>game</main>", "game.ts": "console.log('game')"}})
	if err != nil {
		t.Fatal(err)
	}
	preview, artifact, err := service.Preview(context.Background(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Status != "preview" || artifact.SHA256 == "" {
		t.Fatalf("preview=%+v artifact=%+v", preview, artifact)
	}
	_, archivePath, err := service.Export(context.Background(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	if len(archive.File) != 2 {
		t.Fatalf("archive files=%d", len(archive.File))
	}
	published, publishedPath, err := service.PublishLocal(context.Background(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != "published" {
		t.Fatalf("published=%+v", published)
	}
	if _, err := os.Stat(filepath.Join(publishedPath, "index.html")); err != nil {
		t.Fatal(err)
	}
}

func TestBuilderRejectsTraversal(t *testing.T) {
	service, err := NewBuilderService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), BuilderSpec{Name: "bad", Kind: BuilderWebsite, Files: map[string]string{"../secret": "no"}}); err == nil {
		t.Fatal("path traversal accepted")
	}
}
