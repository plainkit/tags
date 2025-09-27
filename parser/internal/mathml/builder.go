package mathml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"html"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	mathML1URL = "https://www.w3.org/TR/1998/REC-MathML-19980407/appendixF.html"
	mathML2URL = "https://www.w3.org/TR/MathML2/appendixl.html"
	mathML3URL = "https://www.w3.org/TR/MathML3/appendixi.html"
	mathML4URL = "https://www.w3.org/TR/mathml4/"
)

// Output captures the set of MathML tag names collected from the specifications.
type Output struct {
	Meta Meta     `json:"meta"`
	Tags []string `json:"tags"`
}

// Meta summarizes the crawl.
type Meta struct {
	Sources       []string `json:"sources"`
	SchemaVersion string   `json:"schemaVersion"`
	Count         int      `json:"count"`
}

// Build crawls MathML specifications and returns the union of tag names.
func Build(ctx context.Context, client *http.Client, schemaVersion string, seed []string) (Output, error) {
	tagSet := make(map[string]struct{})
	for _, tag := range seed {
		tagSet[strings.TrimSpace(tag)] = struct{}{}
	}

	if err := collectMathML1(ctx, client, tagSet); err != nil {
		return Output{}, err
	}
	if err := collectMathML2(ctx, client, tagSet); err != nil {
		return Output{}, err
	}
	if err := collectMathML3(ctx, client, tagSet); err != nil {
		return Output{}, err
	}
	if err := collectMathML4(ctx, client, tagSet); err != nil {
		return Output{}, err
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	output := Output{
		Meta: Meta{
			Sources:       []string{mathML1URL, mathML2URL, mathML3URL, mathML4URL},
			SchemaVersion: schemaVersion,
			Count:         len(tags),
		},
		Tags: tags,
	}

	return output, nil
}

// Write persists the Output as pretty-printed JSON.
func Write(path string, payload Output) error {
	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mathml index: %w", err)
	}

	jsonBytes = append(jsonBytes, '\n')
	if err := os.WriteFile(path, jsonBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// WriteGo renders the payload as Go source in the provided package.
func WriteGo(path, packageName string, payload Output) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	buf.WriteString("type MathMLMeta struct {\n\tSources []string\n\tSchemaVersion string\n\tCount int\n}\n\n")
	buf.WriteString("type MathMLIndex struct {\n\tMeta MathMLMeta\n\tTags []string\n}\n\n")

	buf.WriteString("var MathML = MathMLIndex{\n")
	buf.WriteString("\tMeta: MathMLMeta{\n")
	buf.WriteString("\t\tSources: []string{\n")
	for _, src := range payload.Meta.Sources {
		fmt.Fprintf(&buf, "\t\t\t%q,\n", src)
	}
	buf.WriteString("\t\t},\n")
	fmt.Fprintf(&buf, "\t\tSchemaVersion: %q,\n", payload.Meta.SchemaVersion)
	fmt.Fprintf(&buf, "\t\tCount: %d,\n", payload.Meta.Count)
	buf.WriteString("\t},\n")

	buf.WriteString("\tTags: []string{\n")
	for _, tag := range payload.Tags {
		fmt.Fprintf(&buf, "\t\t%q,\n", tag)
	}
	buf.WriteString("\t},\n")
	buf.WriteString("}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format mathml go: %w", err)
	}

	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func collectMathML1(ctx context.Context, client *http.Client, tags map[string]struct{}) error {
	doc, err := fetchDocument(ctx, client, mathML1URL)
	if err != nil {
		return err
	}

	links := doc.Find("ul ul ul ul a")
	if links.Length() <= 91 {
		return fmt.Errorf("mathml1: expected more than 91 links, got %d", links.Length())
	}

	links.Each(func(_ int, sel *goquery.Selection) {
		text := html.UnescapeString(strings.TrimSpace(sel.Text()))
		start := strings.Index(text, "<")
		end := strings.Index(text, ">")
		if start == -1 || end == -1 || end <= start+1 {
			return
		}
		value := text[start+1 : end]
		value = strings.TrimSuffix(value, "/")
		value = strings.TrimSpace(value)
		addTag(tags, value)
	})

	return nil
}

func collectMathML2(ctx context.Context, client *http.Client, tags map[string]struct{}) error {
	doc, err := fetchDocument(ctx, client, mathML2URL)
	if err != nil {
		return err
	}

	titles := doc.Find(".div1 .div2:first-child dl dt")
	if titles.Length() <= 188 {
		return fmt.Errorf("mathml2: expected more than 188 titles, got %d", titles.Length())
	}

	titles.Each(func(_ int, sel *goquery.Selection) {
		value := normalize(sel.Text())
		if strings.HasPrefix(value, "m:") {
			return
		}
		addTag(tags, value)
	})

	return nil
}

func collectMathML3(ctx context.Context, client *http.Client, tags map[string]struct{}) error {
	doc, err := fetchDocument(ctx, client, mathML3URL)
	if err != nil {
		return err
	}

	titles := doc.Find(".div1 .div2:first-child dl dt")
	if titles.Length() <= 194 {
		return fmt.Errorf("mathml3: expected more than 194 titles, got %d", titles.Length())
	}

	titles.Each(func(_ int, sel *goquery.Selection) {
		value := normalize(sel.Text())
		switch value {
		case "mi\"", "span":
			return
		}
		addTag(tags, value)
	})

	return nil
}

func collectMathML4(ctx context.Context, client *http.Client, tags map[string]struct{}) error {
	doc, err := fetchDocument(ctx, client, mathML4URL)
	if err != nil {
		return err
	}

	titles := doc.Find("dl#mml_elements dt")
	if titles.Length() <= 179 {
		return fmt.Errorf("mathml4: expected more than 179 titles, got %d", titles.Length())
	}

	titles.Each(func(_ int, sel *goquery.Selection) {
		value := normalize(sel.Text())
		switch {
		case strings.Contains(value, "("), strings.Contains(value, ">"), value == "img":
			return
		}
		addTag(tags, value)
	})

	return nil
}

func addTag(tags map[string]struct{}, raw string) {
	tag := extractTag(raw)
	if tag == "" {
		return
	}
	tags[tag] = struct{}{}
}

func extractTag(raw string) string {
	value := strings.TrimSpace(normalize(raw))
	if value == "" {
		return ""
	}

	if strings.Contains(value, "<") && strings.Contains(value, ">") {
		start := strings.Index(value, "<")
		end := strings.Index(value, ">")
		if start >= 0 && end > start {
			value = value[start+1 : end]
		}
	}

	value = strings.TrimSuffix(value, "/")
	value = strings.Trim(value, "`")
	value = strings.Trim(value, "\"")
	return value
}

func normalize(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00A0", " ")
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "/")
	value = strings.Trim(value, "`")
	value = strings.Trim(value, "\"")
	value = strings.ToLower(value)
	value = strings.TrimSpace(value)
	return value
}

func fetchDocument(ctx context.Context, client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s body: %w", url, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse %s html: %w", url, err)
	}
	return doc, nil
}
