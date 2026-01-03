package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListTopLevelSortTypeDesc(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Mkdir(filepath.Join(tmp, "dirX"), 0755)
	_ = os.WriteFile(filepath.Join(tmp, "fileY"), []byte("y"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{SortBy: "type", OrderDesc: true})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least 2, got %d", len(items))
	}
	if items[0].Type != "file" {
		t.Fatalf("expected file first, got %s", items[0].Type)
	}
}
