package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
)

func teamPage(page *template.Template, lookup func(context.Context) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		data := struct {
			TeamName string
			Error    string
		}{}
		status := http.StatusOK
		name, err := lookup(r.Context())
		if err != nil {
			status = http.StatusServiceUnavailable
			data.Error = "Unable to load the team name from PostgreSQL. Please fix the db connection."
		} else {
			data.TeamName = name
		}
		var body bytes.Buffer
		if err := page.Execute(&body, data); err != nil {
			log.Printf("Page rendering failed: %v", err)
			http.Error(w, "Unable to render page", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write(body.Bytes())
	}
}
