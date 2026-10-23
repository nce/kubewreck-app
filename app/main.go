package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed static/*
var embeddedFiles embed.FS

var version = "dev"

func main() {
	page := template.Must(template.ParseFS(embeddedFiles, "static/index.html"))
	store := newTeamStore(os.Getenv("DATABASE_URL"))
	if store.pool != nil {
		defer store.pool.Close()
		ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
		store.reportConnection(store.pool.Ping(ctx))
		cancel()
	}
	http.HandleFunc("/", teamPage(page, store.teamName))

	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		const upgrade = `
			<p><span style="italic; color:#FFFFC5"><i>Upgrade available:</span></p>
			<div class="code-container">
				kubectl set image deployment/kubewreck-app kubewreck-app=ghcr.io/nce/kubewreck-app:v2
			</div>
		`
		const latest = " 🥳 this this the latest version 🥳"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if version == "v2" {
			fmt.Fprintf(w, latest)
		} else {
			fmt.Fprintf(w, upgrade)
		}
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version)
	})

	log.Printf("Starting kubewreck-app version=%s uid=%d gid=%d", version, os.Getuid(), os.Getgid())
	log.Printf("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
