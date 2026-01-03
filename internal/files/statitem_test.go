package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatItem(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "x.txt")
	_ = os.WriteFile(p, []byte("abc"), 0o644)
	it, err := StatItem(p)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if it.Name != "x.txt" || it.Type != "file" || it.Parent != base {
		t.Fatalf("stat fields wrong: %v %v %v", it.Name, it.Type, it.Parent)
	}
}
