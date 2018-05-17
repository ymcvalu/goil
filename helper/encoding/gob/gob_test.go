package gob

import (
	"testing"
)

type Bean struct {
	Name string
	Age  int
}

func TestGobEncode(t *testing.T) {
	b := Bean{
		"Jim",
		22,
	}
	byts, _ := GobEncode(b)

	b1, _ := GobDecode(byts)
	t.Error(b1.(Bean))

}

func TestGobEncodePointer(t *testing.T) {
	b := Bean{"Jim", 22}
	byts, _ := GobEncode(&b)
	b1, _ := GobDecode(byts)
	t.Error(b1.(*Bean))
}
