package files

import (
	"testing"
)

func TestParseRangeOpenEnded(t *testing.T) {
	var s, e int64
	ok, err := ParseRange("bytes=1-", 6, &s, &e)
	if err != nil || !ok || s != 1 || e != 6 {
		t.Fatalf("unexpected: ok=%v err=%v s=%d e=%d", ok, err, s, e)
	}
}

func TestParseRangeSuffixUnsupported(t *testing.T) {
	var s, e int64
	ok, err := ParseRange("bytes=-3", 6, &s, &e)
	if ok || err == nil {
		t.Fatalf("expected error for suffix range")
	}
}

func TestParseRangeInvalid(t *testing.T) {
	var s, e int64
	ok, err := ParseRange("bytes=-1", 6, &s, &e)
	if ok || err == nil {
		t.Fatalf("expected error")
	}
	ok2, err2 := ParseRange("bytes=2-1", 6, &s, &e)
	if ok2 || err2 == nil {
		t.Fatalf("expected error")
	}
	ok3, err3 := ParseRange("bytes=abc", 6, &s, &e)
	if ok3 || err3 == nil {
		t.Fatalf("expected error")
	}
}
