package files

import (
	"io"
	"net/http"
	"os"
	"strconv"
)

func ServeFile(w http.ResponseWriter, r *http.Request, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()
	w.Header().Set("Accept-Ranges", "bytes")
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return nil
	}
	var start, end int64
	if _, err := ParseRange(rangeHeader, size, &start, &end); err != nil {
		http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
		return nil
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return err
	}
	w.Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end-1, 10)+"/"+strconv.FormatInt(size, 10))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start, 10))
	w.WriteHeader(http.StatusPartialContent)
	_, err = io.CopyN(w, f, end-start)
	return err
}

func ParseRange(h string, size int64, start, end *int64) (bool, error) {
	const prefix = "bytes="
	if len(h) < len(prefix) || h[:len(prefix)] != prefix {
		return false, nil
	}
	spec := h[len(prefix):]
	var s, e int64
	if n, err := strconv.ParseInt(spec, 10, 64); err == nil {
		if n < 0 {
			return false, http.ErrNotSupported
		}
		s = n
		e = size
	} else {
		var dash int
		for i := range spec {
			if spec[i] == '-' {
				dash = i
				break
			}
		}
		if dash == 0 {
			n, err := strconv.ParseInt(spec[1:], 10, 64)
			if err != nil {
				return false, err
			}
			if n > size {
				n = size
			}
			s = size - n
			e = size
		} else {
			left := spec[:dash]
			right := spec[dash+1:]
			n1, err1 := strconv.ParseInt(left, 10, 64)
			if err1 != nil {
				return false, err1
			}
			s = n1
			if right == "" {
				e = size
			} else {
				n2, err2 := strconv.ParseInt(right, 10, 64)
				if err2 != nil {
					return false, err2
				}
				e = n2 + 1
			}
		}
	}
	if s < 0 || e > size || s >= e {
		return false, http.ErrNotSupported
	}
	*start, *end = s, e
	return true, nil
}

