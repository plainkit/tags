package index

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	customElementRE = regexp.MustCompile(`(?i)custom elements`)
	parenRE         = regexp.MustCompile(`\([^)]+\)`)
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

// Build fetches the HTML specification index from specURL, parses attribute metadata,
// and returns a structured Output ready for serialization.
func Build(ctx context.Context, client *http.Client, specURL, schemaVersion string) (Output, error) {
	doc, err := fetchDocument(ctx, client, specURL)
	if err != nil {
		return Output{}, err
	}

	nodes := doc.Find("#attributes-1 tbody tr")
	if nodes.Length() == 0 {
		return Output{}, errors.New("missing results in html")
	}

	attributes := map[string][]string{"*": nil}
	booleanAttrSet := make(map[string]struct{})

	nodes.Each(func(_ int, s *goquery.Selection) {
		cells := s.ChildrenFiltered("td,th")
		if cells.Length() < 2 {
			return
		}

		name := strings.TrimSpace(cells.Eq(0).Text())
		value := strings.TrimSpace(cells.Eq(1).Text())
		if name == "" || value == "" {
			return
		}

		if customElementRE.MatchString(value) {
			return
		}

		if cells.Length() > 3 {
			typeCell := strings.TrimSpace(cells.Eq(3).Text())
			if strings.Contains(strings.ToLower(typeCell), "boolean attribute") {
				booleanAttrSet[name] = struct{}{}
			}
		}

		elements := extractElements(value)
		for _, element := range elements {
			if element == "" {
				continue
			}

			attrList := attributes[element]
			if attrList == nil {
				attrList = []string{}
			}

			if !contains(attrList, name) {
				attrList = append(attrList, name)
			}
			attributes[element] = attrList
		}
	})

	emptyElements := ensureElementCoverage(attributes, doc)
	normalized, globals := normalize(attributes)

	emptySet := make(map[string]struct{}, len(emptyElements))
	for _, name := range emptyElements {
		emptySet[name] = struct{}{}
	}

	attributeSet := make(map[string]struct{})
	globalRefs := make([]AttributeRef, 0, len(globals))
	for _, attr := range globals {
		attributeSet[attr] = struct{}{}
		ref := AttributeRef{Name: attr}
		if _, ok := booleanAttrSet[attr]; ok {
			ref.Boolean = true
		}
		globalRefs = append(globalRefs, ref)
	}

	elementSummaries := make(map[string]ElementSummary, len(normalized))
	for element, attrs := range normalized {
		var refs []AttributeRef
		for _, attr := range attrs {
			attributeSet[attr] = struct{}{}
			ref := AttributeRef{Name: attr}
			if _, ok := booleanAttrSet[attr]; ok {
				ref.Boolean = true
			}
			refs = append(refs, ref)
		}
		_, empty := emptySet[element]
		elementSummaries[element] = ElementSummary{
			Empty:      empty,
			Attributes: refs,
		}
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

func extractElements(value string) []string {
	if strings.Contains(value, "HTML elements") {
		return []string{"*"}
	}

	parts := strings.Split(value, ";")
	elements := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := parenRE.ReplaceAllString(part, "")
		clean = strings.TrimSpace(clean)
		clean = strings.ToLower(clean)
		if clean != "" {
			elements = append(elements, clean)
		}
	}
	return elements
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func ensureElementCoverage(attributes map[string][]string, doc *goquery.Document) []string {
	table := doc.Find("h3#elements-3").NextAllFiltered("table").First()
	if table.Length() == 0 {
		return nil
	}

	empty := make([]string, 0)
	seenEmpty := make(map[string]struct{})

	table.Find("tbody tr").Each(func(_ int, s *goquery.Selection) {
		cells := s.ChildrenFiltered("td")
		hasChildrenInfo := cells.Length() >= 4
		emptyChildren := false
		if hasChildrenInfo {
			childrenCell := strings.TrimSpace(cells.Eq(3).Text())
			emptyChildren = strings.EqualFold(childrenCell, "empty")
		}

		s.Find("th code").Each(func(_ int, codeSel *goquery.Selection) {
			name := strings.TrimSpace(codeSel.Text())
			if name == "" {
				return
			}
			name = strings.ToLower(name)
			if _, ok := attributes[name]; !ok {
				attributes[name] = []string{}
			}

			if emptyChildren {
				if _, ok := seenEmpty[name]; !ok {
					seenEmpty[name] = struct{}{}
					empty = append(empty, name)
				}
			}
		})
	})

	if len(empty) == 0 {
		return nil
	}

	sort.Strings(empty)
	return empty
}

func normalize(attributes map[string][]string) (map[string][]string, []string) {
	globals := sortAndDedupe(attributes["*"])
	attributes["*"] = globals

	globalSet := make(map[string]struct{}, len(globals))
	for _, attr := range globals {
		globalSet[attr] = struct{}{}
	}

	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(map[string][]string, len(keys))
	for _, key := range keys {
		if key == "*" {
			continue
		}

		attrs := sortAndDedupe(attributes[key])
		if len(globalSet) > 0 {
			filtered := make([]string, 0, len(attrs))
			for _, attr := range attrs {
				if _, ok := globalSet[attr]; ok {
					continue
				}
				filtered = append(filtered, attr)
			}
			attrs = filtered
		}

		if len(attrs) == 0 {
			result[key] = nil
			continue
		}

		result[key] = append([]string(nil), attrs...)
	}

	return result, globals
}

func dedupe(values []string) []string {
	if len(values) == 0 {
		return values
	}

	result := values[:1]
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1] {
			result = append(result, values[i])
		}
	}
	return result
}

func sortAndDedupe(values []string) []string {
	if len(values) == 0 {
		return values
	}
	sort.Strings(values)
	return dedupe(values)
}
