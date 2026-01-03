package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndexerWatchingCreatesNewFile(t *testing.T) {
	tmp := t.TempDir()
	ix := NewIndexer()
	ix.SetRoots([]string{tmp})
	_ = ix.Rebuild()
	if err := ix.StartWatching(); err != nil {
		t.Fatalf("start watching error: %v", err)
	}
	defer ix.StopWatching()
	p := filepath.Join(tmp, "created.txt")
	_ = os.WriteFile(p, []byte("x"), 0644)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		items, _ := ix.Search("created.txt", 1, 10)
		if len(items) > 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("new file not indexed in time")
}
