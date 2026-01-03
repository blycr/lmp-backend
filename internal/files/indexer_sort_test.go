package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndexerSearchAdvancedSorts(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	time.Sleep(5 * time.Millisecond)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("bb"), 0644)
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	items1, _ := ix.SearchAdvanced("", SearchOptions{Type: "file", SortBy: "name", OrderDesc: true}, 1, 10)
	if len(items1) != 2 || items1[0].Name != "b.txt" {
		t.Fatalf("expected b.txt first")
	}
	items2, _ := ix.SearchAdvanced("", SearchOptions{SortBy: "size", OrderDesc: true}, 1, 10)
	if items2[0].Name != "b.txt" {
		t.Fatalf("expected larger first")
	}
}
