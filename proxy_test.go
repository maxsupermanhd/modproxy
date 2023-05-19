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
	at, ok := actionHandlers["headers"]
	if !ok {
		t.FailNow()
	}
	w := rw{
		header: map[string][]string{},
	}
	err := at(&w, nil, map[string]interface{}{
		"action": "add",
		"headers": map[string]interface{}{
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
