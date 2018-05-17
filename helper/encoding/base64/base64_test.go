package base64

import (
	"testing"
)

func TestURL(t *testing.T) {
	raw := "this is a raw string"
	src := []byte(raw)
	desc, _ := URLEncoding(src)
	if s, _ := URLDecoding(desc); string(s) != raw {
		t.Error("base64 url error")
	}
}

func TestStd(t *testing.T) {
	raw := "this is a raw string"
	src := []byte(raw)
	desc, _ := StdEncoding(src)
	if s, _ := StdDecoding(desc); string(s) != raw {
		t.Error("base64 std error")
	}
}

func TestRawStd(t *testing.T) {
	raw := "this is a raw string"
	src := []byte(raw)
	desc, _ := RawStdEncoding(src)
	if s, _ := RawStdDecoding(desc); string(s) != raw {
		t.Error("base64 std error")
	}
}
