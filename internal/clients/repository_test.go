package clients

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"testing"
)

func TestClientAndProjectCRUD(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	r := NewRepository(db)
	ctx := context.Background()
	client, err := r.CreateClient(ctx, Client{ID: "client-1", Name: "Empresa ABC"})
	if err != nil {
		t.Fatal(err)
	}
	if client.Name != "Empresa ABC" {
		t.Fatal("client was not created")
	}
	project, err := r.CreateProject(ctx, Project{ID: "project-1", ClientID: client.ID, Name: "Matriz"})
	if err != nil {
		t.Fatal(err)
	}
	items, err := r.ListProjects(ctx, client.ID, "")
	if err != nil || len(items) != 1 {
		t.Fatalf("projects = %d, %v", len(items), err)
	}
	if err := r.DeleteClient(ctx, client.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetProject(ctx, project.ID); err != ErrNotFound {
		t.Fatalf("project should cascade delete: %v", err)
	}
}
