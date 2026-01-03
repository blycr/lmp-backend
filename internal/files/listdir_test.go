package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListDirSortAndFilter(t *testing.T) {
	base := t.TempDir()
	d1 := filepath.Join(base, "a_dir")
	f1 := filepath.Join(base, "b.txt")
	f2 := filepath.Join(base, "c.log")
	_ = os.MkdirAll(d1, 0o755)
	_ = os.WriteFile(f1, []byte("x"), 0o644)
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(f2, []byte("hello"), 0o644)

	items, err := ListDir(base, ListOptions{SortBy: "name"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Name != "a_dir" || items[1].Name != "b.txt" || items[2].Name != "c.log" {
		t.Fatalf("order by name asc wrong: %v %v %v", items[0].Name, items[1].Name, items[2].Name)
	}
	items, _ = ListDir(base, ListOptions{SortBy: "name", OrderDesc: true})
	if items[0].Name != "c.log" || items[2].Name != "a_dir" {
		t.Fatalf("order by name desc wrong: %v %v", items[0].Name, items[2].Name)
	}
	items, _ = ListDir(base, ListOptions{SortBy: "time", OrderDesc: true})
	if items[0].Name != "c.log" {
		t.Fatalf("order by time desc wrong: %v", items[0].Name)
	}
	items, _ = ListDir(base, ListOptions{SortBy: "size"})
	if items[1].Name != "b.txt" || items[2].Name != "c.log" {
		t.Fatalf("order by size asc wrong: %v %v", items[1].Name, items[2].Name)
	}
	items, _ = ListDir(base, ListOptions{SortBy: "type"})
	if items[0].Type != "dir" || items[1].Type != "file" {
		t.Fatalf("order by type wrong: %v %v", items[0].Type, items[1].Type)
	}
	items, _ = ListDir(base, ListOptions{SortBy: "type", OrderDesc: true})
	if items[0].Type != "file" || items[2].Type != "dir" {
		t.Fatalf("order by type desc wrong: %v %v", items[0].Type, items[2].Type)
	}
	items, _ = ListDir(base, ListOptions{Filter: "b."})
	if len(items) != 1 || items[0].Name != "b.txt" {
		t.Fatalf("filter wrong: %d %v", len(items), items[0].Name)
	}
}
