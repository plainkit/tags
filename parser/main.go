package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"parser/internal/html"
	"parser/internal/output"
	"parser/internal/svg"
)

const (
	htmlSpecURL       = "https://html.spec.whatwg.org/multipage/indices.html"
	htmlSchemaVersion = "1.1.0"
	svgSchemaVersion  = "1.0.0"
)

var (
	htmlIndexPath = filepath.Join("..", "json", "html_elements_index.json")
	svgIndexPath  = filepath.Join("..", "json", "svg_elements_index.json")
	htmlGoPath    = filepath.Join("..", "html.go")
	svgGoPath     = filepath.Join("..", "svg.go")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 30 * time.Second}

	if err := output.EnsureDir(htmlIndexPath); err != nil {
		return err
	}
	if err := output.EnsureDir(svgIndexPath); err != nil {
		return err
	}

	htmlOutput, err := html.Build(ctx, client, htmlSpecURL, htmlSchemaVersion)
	if err != nil {
		return err
	}
	if err := html.Write(htmlIndexPath, htmlOutput); err != nil {
		return err
	}
	if err := html.WriteGo(htmlGoPath, "tags", htmlOutput); err != nil {
		return err
	}

	svgOutput, err := svg.Build(ctx, client, svgSchemaVersion)
	if err != nil {
		return err
	}
	if err := svg.Write(svgIndexPath, svgOutput); err != nil {
		return err
	}
	if err := svg.WriteGo(svgGoPath, "tags", svgOutput); err != nil {
		return err
	}

	return nil
}
