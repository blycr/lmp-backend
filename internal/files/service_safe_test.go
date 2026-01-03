package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSafePathValid(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "a.txt")
	_ = os.WriteFile(p, []byte("x"), 0644)
	got, ok := ResolveSafePath([]string{root}, p)
	if !ok || got != p {
		t.Fatalf("expected valid path")
	}
}

func TestResolveSafePathInvalid(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "x.txt")
	_ = os.WriteFile(outside, []byte("x"), 0644)
	_, ok := ResolveSafePath([]string{root}, outside)
	if ok {
		t.Fatalf("expected invalid path")
	}
}
