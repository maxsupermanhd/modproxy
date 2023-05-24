package main

import "net/http"

type loggingResponseWriter struct {
	http.ResponseWriter
	ResponseCode *int
}

func (l loggingResponseWriter) WriteHeader(statusCode int) {
	*l.ResponseCode = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}
