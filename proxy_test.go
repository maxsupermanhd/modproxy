package main

import (
	"net/http"
	"testing"
)

type rw struct {
	header http.Header
}

func (r *rw) Header() http.Header {
	return r.header
}
func (*rw) WriteHeader(statusCode int) {

}
func (*rw) Write([]byte) (int, error) {
	return 0, nil
}

func TestActionHeaders(t *testing.T) {
	at, ok := requestActionHandlers["headers"]
	if !ok {
		t.FailNow()
	}
	w := rw{
		header: map[string][]string{},
	}
	err := at(&w, nil, map[string]any{
		"action": "add",
		"headers": map[string]any{
			"funny": "header",
		},
	})
	if err != nil {
		t.FailNow()
	}
	if w.Header().Get("funny") != "header" {
		t.FailNow()
	}
}
