package files

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileItem struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	ModTime time.Time `json:"mod_time"`
	Parent  string    `json:"parent"`
}

type ListOptions struct {
	Filter    string
	SortBy    string // name|size|time|type
	OrderDesc bool
}

func ListTopLevel(dirs []string, opt ListOptions) ([]FileItem, error) {
	var out []FileItem
	for _, d := range dirs {
		if d == "" {
			continue
		}
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			full := filepath.Join(d, e.Name())
			fi, err := os.Stat(full)
			if err != nil {
				continue
			}
			item := FileItem{
				Name:    e.Name(),
				Path:    full,
				Size:    fi.Size(),
				Type:    fileType(fi),
				ModTime: fi.ModTime(),
				Parent:  d,
			}
			if opt.Filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(opt.Filter)) {
				continue
			}
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		switch opt.SortBy {
		case "size":
			if opt.OrderDesc {
				return out[i].Size > out[j].Size
			}
			return out[i].Size < out[j].Size
		case "time":
			if opt.OrderDesc {
				return out[i].ModTime.After(out[j].ModTime)
			}
			return out[i].ModTime.Before(out[j].ModTime)
		case "type":
			if opt.OrderDesc {
				return out[i].Type > out[j].Type
			}
			return out[i].Type < out[j].Type
		default:
			if opt.OrderDesc {
				return out[i].Name > out[j].Name
			}
			return out[i].Name < out[j].Name
		}
	})
	return out, nil
}

func fileType(fi os.FileInfo) string {
	if fi.IsDir() {
		return "dir"
	}
	return "file"
}

func ResolveSafePath(roots []string, req string) (string, bool) {
	if req == "" {
		return "", false
	}
	clean := filepath.Clean(req)
	for _, r := range roots {
		if r == "" {
			continue
		}
		base := filepath.Clean(r)
		rel, err := filepath.Rel(base, clean)
		if err != nil {
			continue
		}
		if rel == "." || (!strings.HasPrefix(rel, "..") && !strings.Contains(rel, string(filepath.Separator)+"..")) {
			return clean, true
		}
	}
	return "", false
}

func ListDir(dir string, opt ListOptions) ([]FileItem, error) {
	var out []FileItem
	if dir == "" {
		return out, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out, nil
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		fi, err := os.Stat(full)
		if err != nil {
			continue
		}
		item := FileItem{
			Name:    e.Name(),
			Path:    full,
			Size:    fi.Size(),
			Type:    fileType(fi),
			ModTime: fi.ModTime(),
			Parent:  dir,
		}
		if opt.Filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(opt.Filter)) {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		switch opt.SortBy {
		case "size":
			if opt.OrderDesc {
				return out[i].Size > out[j].Size
			}
			return out[i].Size < out[j].Size
		case "time":
			if opt.OrderDesc {
				return out[i].ModTime.After(out[j].ModTime)
			}
			return out[i].ModTime.Before(out[j].ModTime)
		case "type":
			if opt.OrderDesc {
				return out[i].Type > out[j].Type
			}
			return out[i].Type < out[j].Type
		default:
			if opt.OrderDesc {
				return out[i].Name > out[j].Name
			}
			return out[i].Name < out[j].Name
		}
	})
	return out, nil
}

func StatItem(path string) (FileItem, error) {
	var it FileItem
	fi, err := os.Stat(path)
	if err != nil {
		return it, err
	}
	it = FileItem{
		Name:    filepath.Base(path),
		Path:    path,
		Size:    fi.Size(),
		Type:    fileType(fi),
		ModTime: fi.ModTime(),
		Parent:  filepath.Dir(path),
	}
	return it, nil
}
