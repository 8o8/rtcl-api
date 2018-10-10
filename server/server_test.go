package server_test

import (
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/server"
)

// TestNewServer creates a new server and tests the root path for a 200 ok response
func TestNewServer(t *testing.T) {
	is := is.New(t)

	cfg := server.Config{
		Port: "8888",
		Token: server.TokenConfig{
			Issuer:     "Test Issuer",
			SigningKey: "SomeRandomSigningKey",
			HoursTTL:   4,
		},
	}
	s := server.NewServer(cfg, nil)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // expected 200 ok
}
