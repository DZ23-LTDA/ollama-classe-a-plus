package agent

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// §7: rollback de versao publicada — publica v1, edita e publica v2, faz
// rollback para v1, e cobre cross-tenant + versao inexistente.
func TestBuilderRollbackPublishToPreviousVersion(t *testing.T) {
	service, err := NewBuilderService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	project, err := service.Create(ctx, BuilderSpec{Name: "Site", Kind: BuilderWebsite, Components: []VisualComponent{{ID: "hero", Type: "hero", Props: map[string]string{"text": "v1"}}}})
	if err != nil {
		t.Fatal(err)
	}
	// Publica v1.
	if _, p1, err := service.PublishLocal(ctx, project.ID); err != nil || !strings.HasSuffix(p1, "v1") {
		t.Fatalf("publish v1: path=%q err=%v", p1, err)
	}
	// Edita (Version -> 2) e publica v2.
	if _, err := service.ApplyVisualComponents(ctx, project.ID, []VisualComponent{{ID: "button", Type: "button", Props: map[string]string{"text": "v2"}}}); err != nil {
		t.Fatal(err)
	}
	if _, p2, err := service.PublishLocal(ctx, project.ID); err != nil || !strings.HasSuffix(p2, "v2") {
		t.Fatalf("publish v2: path=%q err=%v", p2, err)
	}

	// Rollback para v1.
	rolled, path, err := service.RollbackPublish(ctx, project.ID, project.OrganizationID, 1)
	if err != nil {
		t.Fatalf("rollback v1 err=%v", err)
	}
	if !strings.HasSuffix(path, "v1") || !strings.HasSuffix(rolled.PublishedPath, "v1") || rolled.Status != "published" {
		t.Fatalf("rollback nao apontou para v1: path=%q project=%+v", path, rolled)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("dir da v1 deveria existir: %v", err)
	}

	// Cross-tenant: outra org nao pode fazer rollback.
	if _, _, err := service.RollbackPublish(ctx, project.ID, "org-intruso", 1); !errors.Is(err, ErrBuilderForbidden) {
		t.Fatalf("cross-tenant rollback esperava forbidden, got %v", err)
	}

	// Versao nunca publicada.
	if _, _, err := service.RollbackPublish(ctx, project.ID, project.OrganizationID, 99); !errors.Is(err, ErrBuilderVersionNotPublished) {
		t.Fatalf("rollback p/ versao inexistente esperava not-published, got %v", err)
	}
}
