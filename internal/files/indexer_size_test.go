package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexerSearchAdvancedSizeRange(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "bb.txt"), []byte("bb"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "ccc.txt"), []byte("ccc"), 0644)
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	items, total := ix.SearchAdvanced("", SearchOptions{MinSize: 2, MaxSize: 2}, 1, 10)
	if total != 1 || len(items) != 1 || items[0].Name != "bb.txt" {
		t.Fatalf("expected only bb.txt")
	}
}
