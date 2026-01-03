package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchAdvancedFilters(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "small.bin"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "large.bin"), make([]byte, 2048), 0644)
	_ = os.Mkdir(filepath.Join(tmp, "dir"), 0755)
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	items1, total1 := ix.SearchAdvanced("", SearchOptions{Type: "dir"}, 1, 10)
	if len(items1) == 0 || total1 == 0 {
		t.Fatalf("expected some dirs")
	}
	items2, _ := ix.SearchAdvanced("", SearchOptions{Type: "file", MinSize: 1024}, 1, 10)
	if len(items2) == 0 {
		t.Fatalf("expected large file matched")
	}
	items3, _ := ix.SearchAdvanced("", SearchOptions{Type: "FiLe"}, 1, 10)
	if len(items3) == 0 {
		t.Fatalf("expected case-insensitive type match")
	}
}
