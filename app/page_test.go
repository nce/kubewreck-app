package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTeamPage(t *testing.T) {
	page := template.Must(template.ParseFS(embeddedFiles, "static/index.html"))
	for _, tc := range []struct {
		name   string
		team   string
		err    error
		status int
		want   string
	}{
		{"database name", "opdt", nil, http.StatusOK, "Hello opdt"},
		{"escaped name", "<script>alert(1)</script>", nil, http.StatusOK, "Hello &lt;script&gt;"},
		{"database failure", "", errors.New("private database details"), http.StatusServiceUnavailable, "Unable to load the team name from PostgreSQL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := teamPage(page, func(context.Context) (string, error) { return tc.team, tc.err })
			for _, path := range []string{"/", "/index.html"} {
				w := httptest.NewRecorder()
				handler(w, httptest.NewRequest(http.MethodGet, path, nil))
				if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.want) {
					t.Fatalf("%s: status=%d body=%s", path, w.Code, w.Body.String())
				}
				if strings.Contains(w.Body.String(), "private database details") {
					t.Fatal("database error details leaked to page")
				}
				if !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
					t.Fatal("team page must not be cached")
				}
			}
		})
	}
}

func TestInvalidDatabaseConfiguration(t *testing.T) {
	t.Setenv("PGHOST", "")
	t.Setenv("POSTGRES_SECRET_DIR", "")
	for _, url := range []string{"", "postgres://%invalid"} {
		store := newTeamStore(url)
		if _, err := store.teamName(context.Background()); err == nil {
			t.Fatal("expected lookup to fail for invalid configuration")
		}
	}
}
