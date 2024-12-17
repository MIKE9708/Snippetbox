package main

import(
	"bytes"
	"io"
    "net/http"
    "net/http/httptest"
    "testing"
	"snippetbox.mike9708.net/internal/assert"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("Ok"))

	}) 
	
	secureHeaders(next).ServeHTTP(rr, r)
	rs := rr.Result()

	expected_values := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expected_values)

	expected_values = "origin-when-cross-origin"
    assert.Equal(t, rs.Header.Get("Referrer-Policy"), expected_values)

	expected_values = "nosniff"
    assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expected_values)
	
	expected_values = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expected_values)

	expected_values = "0"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expected_values)

	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer rs.Body.Close()
		body, err := io.ReadAll(rs.Body) 
		if err != nil {
			t.Fatal(err)
		}
	bytes.TrimSpace(body)

	assert.Equal(t, string(body), "Ok")
} 
