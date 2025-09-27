package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"parser/internal/html"
)

const (
	specURL           = "https://html.spec.whatwg.org/multipage/indices.html"
	elementsIndexPath = "html_elements_index.json"
	schemaVersion     = "1.1.0"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	output, err := html.Build(ctx, client, specURL, schemaVersion)
	if err != nil {
		return err
	}

	return html.Write(elementsIndexPath, output)
}
