package main

import (
	"testing"
	"snippetbox.mike9708.net/internal/assert"
    "net/http"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes()) 
	defer ts.Close()

	code, _, body:= ts.get(t, "/ping") 

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, string(body), "Ok")
}
