package goil

import (
	"testing"
)

func TestJoinPath(t *testing.T) {
	url := joinPath("v1", "bean")
	if url != "v1/bean" {
		t.Error(url)
	}
	url = joinPath("v1/", "/bean/")
	if url != "v1/bean/" {
		t.Error(url)
	}
	url = joinPath("/v1/", "/bean/")
	if url != "/v1/bean/" {
		t.Error(url)
	}
}
