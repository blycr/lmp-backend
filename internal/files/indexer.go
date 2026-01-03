package files

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Indexer struct {
	mu      sync.RWMutex
	roots   []string
	items   []FileItem
	watcher *fsnotify.Watcher
	cancel  context.CancelFunc
}

func NewIndexer() *Indexer {
	return &Indexer{}
}

func (ix *Indexer) SetRoots(roots []string) {
	var out []string
	seen := make(map[string]struct{})
	for _, r := range roots {
		if r == "" {
			continue
		}
		c := filepath.Clean(r)
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	ix.mu.Lock()
	ix.roots = out
	ix.mu.Unlock()
}

func (ix *Indexer) Rebuild() error {
	var collected []FileItem
	for _, root := range ix.currentRoots() {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			fi, err := os.Stat(path)
			if err != nil {
				return nil
			}
			item := FileItem{
				Name:    d.Name(),
				Path:    path,
				Size:    fi.Size(),
				Type:    fileType(fi),
				ModTime: fi.ModTime(),
				Parent:  root,
			}
			collected = append(collected, item)
			return nil
		})
	}
	sort.Slice(collected, func(i, j int) bool {
		return strings.ToLower(collected[i].Name) < strings.ToLower(collected[j].Name)
	})
	ix.mu.Lock()
	ix.items = collected
	ix.mu.Unlock()
	return nil
}

func (ix *Indexer) currentRoots() []string {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	out := make([]string, len(ix.roots))
	copy(out, ix.roots)
	return out
}

func (ix *Indexer) Search(q string, page, size int) ([]FileItem, int) {
	qq := strings.ToLower(strings.TrimSpace(q))
	if size <= 0 {
		size = 50
	}
	var filtered []FileItem
	ix.mu.RLock()
	for _, it := range ix.items {
		if qq == "" || strings.Contains(strings.ToLower(it.Name), qq) {
			filtered = append(filtered, it)
		}
	}
	ix.mu.RUnlock()
	total := len(filtered)
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

type SearchOptions struct {
	Type           string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
	SortBy         string
	OrderDesc      bool
}

func (ix *Indexer) SearchAdvanced(q string, opt SearchOptions, page, size int) ([]FileItem, int) {
	qq := strings.ToLower(strings.TrimSpace(q))
	if size <= 0 {
		size = 50
	}
	var filtered []FileItem
	ix.mu.RLock()
	for _, it := range ix.items {
		if qq != "" && !strings.Contains(strings.ToLower(it.Name), qq) {
			continue
		}
		if opt.Type != "" && !strings.EqualFold(opt.Type, it.Type) {
			continue
		}
		if opt.MinSize > 0 && it.Size < opt.MinSize {
			continue
		}
		if opt.MaxSize > 0 && it.Size > opt.MaxSize {
			continue
		}
		if !opt.ModifiedAfter.IsZero() && it.ModTime.Before(opt.ModifiedAfter) {
			continue
		}
		if !opt.ModifiedBefore.IsZero() && it.ModTime.After(opt.ModifiedBefore) {
			continue
		}
		filtered = append(filtered, it)
	}
	ix.mu.RUnlock()
	sort.Slice(filtered, func(i, j int) bool {
		switch opt.SortBy {
		case "size":
			if opt.OrderDesc {
				return filtered[i].Size > filtered[j].Size
			}
			return filtered[i].Size < filtered[j].Size
		case "time":
			if opt.OrderDesc {
				return filtered[i].ModTime.After(filtered[j].ModTime)
			}
			return filtered[i].ModTime.Before(filtered[j].ModTime)
		case "type":
			if opt.OrderDesc {
				return filtered[i].Type > filtered[j].Type
			}
			return filtered[i].Type < filtered[j].Type
		default:
			if opt.OrderDesc {
				return strings.ToLower(filtered[i].Name) > strings.ToLower(filtered[j].Name)
			}
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		}
	})
	total := len(filtered)
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

var defaultIndexer *Indexer

func SetDefaultIndexer(ix *Indexer) {
	defaultIndexer = ix
}

func GetDefaultIndexer() *Indexer {
	return defaultIndexer
}

func (ix *Indexer) StartWatching() error {
	ix.StopWatching()
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	roots := ix.currentRoots()
	for _, r := range roots {
		filepath.WalkDir(r, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				_ = w.Add(path)
			}
			return nil
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	ix.mu.Lock()
	ix.watcher = w
	ix.cancel = cancel
	ix.mu.Unlock()
	go func() {
		defer w.Close()
		var last time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-w.Events:
				if ev.Op&fsnotify.Create == fsnotify.Create {
					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						_ = w.Add(ev.Name)
					}
				}
				if time.Since(last) > 300*time.Millisecond {
					_ = ix.Rebuild()
					last = time.Now()
				}
			case <-w.Errors:
				if time.Since(last) > 300*time.Millisecond {
					_ = ix.Rebuild()
					last = time.Now()
				}
			}
		}
	}()
	return nil
}

func (ix *Indexer) StopWatching() {
	ix.mu.Lock()
	c := ix.cancel
	ix.cancel = nil
	w := ix.watcher
	ix.watcher = nil
	ix.mu.Unlock()
	if c != nil {
		c()
	}
	if w != nil {
		_ = w.Close()
	}
}
