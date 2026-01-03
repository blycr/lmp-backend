package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexerSearchBasic(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "alpha.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "beta.txt"), []byte("b"), 0644)
	_ = os.Mkdir(filepath.Join(tmp, "folder"), 0755)
	_ = os.WriteFile(filepath.Join(tmp, "folder", "gamma.txt"), []byte("g"), 0644)
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	items, total := ix.Search("a", 1, 10)
	if total < 2 {
		t.Fatalf("expected at least 2 results, got %d", total)
	}
	if len(items) == 0 {
		t.Fatalf("expected some items")
	}
}

func TestIndexerSearchPaging(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("b"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "c.txt"), []byte("c"), 0644)
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	p1, _ := ix.Search("", 1, 2)
	if len(p1) != 2 {
		t.Fatalf("expected page size 2, got %d", len(p1))
	}
	p2, _ := ix.Search("", 2, 2)
	if len(p2) < 1 {
		t.Fatalf("expected second page to have items")
	}
}
