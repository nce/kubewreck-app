package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var embeddedFiles embed.FS

var version = "dev"

func main() {
	subFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(subFS))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xhtml+xml; charset=utf-8")
		fileServer.ServeHTTP(w, r)
	})

	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		const upgrade = `
			<p><span style="italic; color:#FFFFC5"><i>Upgrade available:</span></p>
			<div class="code-container">
				kubectl set image deployment/kubewreck kubewreck=ghcr.io/nce/kubewreck:v2
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

	log.Printf("Listening on :8080 with version: %s\n", version)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
