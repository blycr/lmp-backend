package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListTopLevelSortTimeDesc(t *testing.T) {
	tmp := t.TempDir()
	a := filepath.Join(tmp, "a.txt")
	b := filepath.Join(tmp, "b.txt")
	_ = os.WriteFile(a, []byte("a"), 0644)
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(b, []byte("b"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{SortBy: "time", OrderDesc: true})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].Name != "b.txt" {
		t.Fatalf("expected b.txt first, got %s", items[0].Name)
	}
}
