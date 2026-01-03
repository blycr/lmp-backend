package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListTopLevelBasic(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("bb"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{SortBy: "name"})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].Name != "a.txt" {
		t.Fatalf("expected a.txt first, got %s", items[0].Name)
	}
}

func TestListTopLevelFilterSort(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "x.txt"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "y.log"), []byte("y"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "z.txt"), []byte("zz"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{Filter: ".txt", SortBy: "size", OrderDesc: true})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].Name != "z.txt" {
		t.Fatalf("expected z.txt first, got %s", items[0].Name)
	}
}

func TestListTopLevelSortTypeAsc(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Mkdir(filepath.Join(tmp, "dirA"), 0755)
	_ = os.WriteFile(filepath.Join(tmp, "fileB"), []byte("b"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{SortBy: "type", OrderDesc: false})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least 2, got %d", len(items))
	}
	if items[0].Type != "dir" {
		t.Fatalf("expected dir first, got %s", items[0].Type)
	}
}

func TestListTopLevelSortNameDesc(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("b"), 0644)
	items, err := ListTopLevel([]string{tmp}, ListOptions{SortBy: "name", OrderDesc: true})
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
