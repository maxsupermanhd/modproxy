package main

import "testing"

func TestMonitor(t *testing.T) {
	c := make(chan logRequest, 2)
	mon.newConns <- c
	l := logRequest{
		Err:            "test1",
		ResponseCode:   420,
		UserAgent:      "tes2t",
		Path:           "te3st",
		SourceIP:       "t4est",
		When:           "5test",
		ProcessingTime: "6",
	}
	mon.e <- l
	ll := <-c
	if ll != l {
		t.FailNow()
	}
}
