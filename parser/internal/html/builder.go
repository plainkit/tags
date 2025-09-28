package html

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Output captures the indexed HTML element metadata produced from the WHATWG specification.
// It is designed for downstream tooling that needs to reason about element-specific attributes.
type Output struct {
	Meta     OutputMeta                `json:"meta"`
	Globals  []AttributeRef            `json:"globals"`
	Elements map[string]ElementSummary `json:"elements"`
}

// OutputMeta summarizes the build that generated the index.
type OutputMeta struct {
	Source         string `json:"source"`
	SchemaVersion  string `json:"schemaVersion"`
	ElementCount   int    `json:"elementCount"`
	AttributeCount int    `json:"attributeCount"`
}

// ElementSummary describes a single HTML element.
type ElementSummary struct {
	Empty      bool           `json:"empty"`
	Attributes []AttributeRef `json:"attributes,omitempty"`
}

// AttributeRef references an attribute by name along with supplemental metadata.
type AttributeRef struct {
	Name    string `json:"name"`
	Boolean bool   `json:"boolean,omitempty"`
}

type elementRow struct {
	empty      bool
	hasGlobals bool
	attributes []string
}

// Build fetches the HTML specification index from specURL, parses attribute metadata,
// and returns a structured Output ready for serialization.
func Build(ctx context.Context, client *http.Client, specURL, schemaVersion string) (Output, error) {
	doc, err := fetchDocument(ctx, client, specURL)
	if err != nil {
		return Output{}, err
	}

	booleanAttrs, globalAttrs, attributeElements, err := parseAttributeMetadata(doc)
	if err != nil {
		return Output{}, err
	}

	elements, err := parseElementsTable(doc)
	if err != nil {
		return Output{}, err
	}

	for elementName, attrs := range attributeElements {
		row, ok := elements[elementName]
		if !ok {
			continue
		}
		row.attributes = append(row.attributes, attrs...)
		elements[elementName] = row
	}

	globalAttrSet := make(map[string]struct{}, len(globalAttrs))
	for _, attr := range globalAttrs {
		globalAttrSet[attr] = struct{}{}
	}

	globalRefs := make([]AttributeRef, 0, len(globalAttrs))
	for _, attr := range globalAttrs {
		ref := AttributeRef{Name: attr}
		if _, ok := booleanAttrs[attr]; ok {
			ref.Boolean = true
		}
		globalRefs = append(globalRefs, ref)
	}

	attributeSet := make(map[string]struct{}, len(globalAttrs))
	for _, attr := range globalAttrs {
		attributeSet[attr] = struct{}{}
	}

	elementSummaries := make(map[string]ElementSummary, len(elements))
	for name, row := range elements {
		filtered := make([]string, 0, len(row.attributes))
		seen := make(map[string]struct{}, len(row.attributes))
		for _, attr := range row.attributes {
			attr = strings.ToLower(strings.TrimSpace(attr))
			if attr == "" {
				continue
			}
			if _, ok := seen[attr]; ok {
				continue
			}
			seen[attr] = struct{}{}
			if strings.HasPrefix(attr, "on") {
				continue
			}
			if row.hasGlobals {
				if _, ok := globalAttrSet[attr]; ok {
					continue
				}
			}
			filtered = append(filtered, attr)
		}
		sort.Strings(filtered)

		refs := make([]AttributeRef, 0, len(filtered))
		for _, attr := range filtered {
			attributeSet[attr] = struct{}{}
			ref := AttributeRef{Name: attr}
			if _, ok := booleanAttrs[attr]; ok {
				ref.Boolean = true
			}
			refs = append(refs, ref)
		}

		summary := ElementSummary{Empty: row.empty}
		if len(refs) > 0 {
			summary.Attributes = refs
		}
		elementSummaries[name] = summary
	}

	attributeCount := len(attributeSet)

	return Output{
		Meta: OutputMeta{
			Source:         specURL,
			SchemaVersion:  schemaVersion,
			ElementCount:   len(elementSummaries),
			AttributeCount: attributeCount,
		},
		Globals:  globalRefs,
		Elements: elementSummaries,
	}, nil
}

// Write persists the Output as pretty-printed JSON at the provided filesystem path.
func Write(path string, payload Output) error {
	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal html element index: %w", err)
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

	buf.WriteString("type AttributeRef struct {\n\tName string\n\tBoolean bool\n}\n\n")
	buf.WriteString("type HTMLElement struct {\n\tEmpty bool\n\tAttributes []AttributeRef\n}\n\n")
	buf.WriteString("type HTMLMeta struct {\n\tSource string\n\tSchemaVersion string\n\tElementCount int\n\tAttributeCount int\n}\n\n")
	buf.WriteString("type HTMLIndex struct {\n\tMeta HTMLMeta\n\tGlobals []AttributeRef\n\tElements map[string]HTMLElement\n}\n\n")

	buf.WriteString("var HTML = HTMLIndex{\n")
	buf.WriteString("\tMeta: HTMLMeta{\n")
	fmt.Fprintf(&buf, "\t\tSource: %q,\n", payload.Meta.Source)
	fmt.Fprintf(&buf, "\t\tSchemaVersion: %q,\n", payload.Meta.SchemaVersion)
	fmt.Fprintf(&buf, "\t\tElementCount: %d,\n", payload.Meta.ElementCount)
	fmt.Fprintf(&buf, "\t\tAttributeCount: %d,\n", payload.Meta.AttributeCount)
	buf.WriteString("\t},\n")

	buf.WriteString("\tGlobals: []AttributeRef{\n")
	for _, ref := range payload.Globals {
		if ref.Boolean {
			fmt.Fprintf(&buf, "\t\t{Name: %q, Boolean: true},\n", ref.Name)
		} else {
			fmt.Fprintf(&buf, "\t\t{Name: %q},\n", ref.Name)
		}
	}
	buf.WriteString("\t},\n")

	keys := make([]string, 0, len(payload.Elements))
	for key := range payload.Elements {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	buf.WriteString("\tElements: map[string]HTMLElement{\n")
	for _, key := range keys {
		element := payload.Elements[key]
		fmt.Fprintf(&buf, "\t\t%q: {\n", key)
		if element.Empty {
			buf.WriteString("\t\t\tEmpty: true,\n")
		}
		if len(element.Attributes) > 0 {
			buf.WriteString("\t\t\tAttributes: []AttributeRef{\n")
			for _, attr := range element.Attributes {
				if attr.Boolean {
					fmt.Fprintf(&buf, "\t\t\t\t{Name: %q, Boolean: true},\n", attr.Name)
				} else {
					fmt.Fprintf(&buf, "\t\t\t\t{Name: %q},\n", attr.Name)
				}
			}
			buf.WriteString("\t\t\t},\n")
		}
		buf.WriteString("\t\t},\n")
	}
	buf.WriteString("\t},\n")
	buf.WriteString("}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format html go: %w", err)
	}

	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func parseAttributeMetadata(doc *goquery.Document) (map[string]struct{}, []string, map[string][]string, error) {
	rows := doc.Find("#attributes-1 tbody tr")
	if rows.Length() == 0 {
		return nil, nil, nil, errors.New("missing attribute metadata in html")
	}

	booleanAttrs := make(map[string]struct{})
	globalSet := make(map[string]struct{})
	elementAttrs := make(map[string][]string)

	rows.Each(func(_ int, s *goquery.Selection) {
		cells := s.ChildrenFiltered("td,th")
		if cells.Length() < 4 {
			return
		}

		name := strings.ToLower(strings.TrimSpace(cells.Eq(0).Text()))
		if name == "" {
			return
		}

		typeCell := strings.ToLower(strings.TrimSpace(cells.Eq(3).Text()))
		if strings.Contains(typeCell, "boolean attribute") {
			booleanAttrs[name] = struct{}{}
		}

		elementCell := cells.Eq(1)
		if strings.Contains(strings.ToLower(elementCell.Text()), "html elements") {
			globalSet[name] = struct{}{}
		}

		names := collectElementNames(elementCell)
		if len(names) == 0 {
			return
		}

		for _, el := range names {
			if el == "" {
				continue
			}
			elementAttrs[el] = append(elementAttrs[el], name)
		}
	})

	globals := make([]string, 0, len(globalSet))
	for attr := range globalSet {
		globals = append(globals, attr)
	}
	sort.Strings(globals)

	return booleanAttrs, globals, elementAttrs, nil
}

func parseElementsTable(doc *goquery.Document) (map[string]elementRow, error) {
	table := doc.Find("h3#elements-3").NextAllFiltered("table").First()
	if table.Length() == 0 {
		return nil, errors.New("missing element metadata in html")
	}

	result := make(map[string]elementRow)
	var parseErr error

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		if parseErr != nil {
			return
		}

		cells := row.ChildrenFiltered("th,td")
		if cells.Length() < 6 {
			parseErr = errors.New("unexpected element row shape in html table")
			return
		}

		names := collectElementNames(cells.Eq(0))
		if len(names) == 0 {
			return
		}

		entry := elementRow{}
		childrenText := strings.TrimSpace(cells.Eq(4).Text())
		entry.empty = strings.EqualFold(childrenText, "empty")

		attrCell := cells.Eq(5)
		entry.hasGlobals = attrCell.Find("a[href$='#global-attributes']").Length() > 0
		entry.attributes = collectAttributeNames(attrCell)

		for _, name := range names {
			result[name] = entry
		}
	})

	if parseErr != nil {
		return nil, parseErr
	}

	if len(result) == 0 {
		return nil, errors.New("parsed zero elements from html table")
	}

	return result, nil
}

func collectElementNames(cell *goquery.Selection) []string {
	if cell == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var names []string

	cell.Find("code").Each(func(_ int, sel *goquery.Selection) {
		name := strings.ToLower(strings.TrimSpace(sel.Text()))
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	})

	return names
}

func collectAttributeNames(cell *goquery.Selection) []string {
	if cell == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var names []string

	cell.Find("code").Each(func(_ int, sel *goquery.Selection) {
		name := strings.ToLower(strings.TrimSpace(sel.Text()))
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	})

	if len(names) == 0 {
		return nil
	}

	sort.Strings(names)
	return names
}

func fetchDocument(ctx context.Context, client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch spec: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch spec: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read spec body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse spec html: %w", err)
	}

	return doc, nil
}
