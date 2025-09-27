package svg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	svg11URL   = "https://www.w3.org/TR/SVG11/attindex.html"
	svgTinyURL = "https://www.w3.org/TR/SVGTiny12/attributeTable.html"
	svg2URL    = "https://www.w3.org/TR/SVG2/attindex.html"
)

// Output captures the indexed SVG element metadata.
type Output struct {
	Meta     OutputMeta                `json:"meta"`
	Globals  []string                  `json:"globals"`
	Elements map[string]ElementSummary `json:"elements"`
}

// OutputMeta summarizes the SVG index build.
type OutputMeta struct {
	Sources        []string `json:"sources"`
	SchemaVersion  string   `json:"schemaVersion"`
	ElementCount   int      `json:"elementCount"`
	AttributeCount int      `json:"attributeCount"`
}

// ElementSummary describes an SVG element and its non-global attributes.
type ElementSummary struct {
	Attributes []string `json:"attributes,omitempty"`
}

// Build constructs an SVG element attribute index by crawling the SVG specifications.
func Build(ctx context.Context, client *http.Client, schemaVersion string) (Output, error) {
	svg11, err := fetchSVG11(ctx, client)
	if err != nil {
		return Output{}, err
	}

	ensureSymbolDefaults(svg11)

	svgTiny, err := fetchSVGTiny(ctx, client)
	if err != nil {
		return Output{}, err
	}

	svg2, err := fetchSVG2(ctx, client)
	if err != nil {
		return Output{}, err
	}

	maps := []elementMap{svg11, svgTiny, svg2}

	globals := collectGlobals(maps)
	merged := mergeMaps(maps, globals)

	globalList := setToSortedSlice(globals)

	elements := make(map[string]ElementSummary, len(merged))
	attributeSet := make(map[string]struct{}, len(globalList))
	for _, attr := range globalList {
		attributeSet[attr] = struct{}{}
	}

	for element, attrs := range merged {
		sorted := setToSortedSlice(attrs)
		if len(sorted) > 0 {
			elements[element] = ElementSummary{Attributes: sorted}
			for _, attr := range sorted {
				attributeSet[attr] = struct{}{}
			}
		} else {
			elements[element] = ElementSummary{}
		}
	}

	return Output{
		Meta: OutputMeta{
			Sources:        []string{svg11URL, svgTinyURL, svg2URL},
			SchemaVersion:  schemaVersion,
			ElementCount:   len(elements),
			AttributeCount: len(attributeSet),
		},
		Globals:  globalList,
		Elements: elements,
	}, nil
}

// Write persists the Output as pretty-printed JSON.
func Write(path string, payload Output) error {
	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal svg element index: %w", err)
	}

	jsonBytes = append(jsonBytes, '\n')
	if err := os.WriteFile(path, jsonBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

type elementMap map[string]map[string]struct{}

func fetchSVG11(ctx context.Context, client *http.Client) (elementMap, error) {
	doc, err := fetchDocument(ctx, client, svg11URL)
	if err != nil {
		return nil, err
	}

	rows := doc.Find(".property-table tr")
	if rows.Length() == 0 {
		return nil, errors.New("could not find rows in SVG 1.1 index")
	}

	result := make(elementMap)
	rows.Each(func(_ int, row *goquery.Selection) {
		attributes := row.Find(".attr-name")
		elements := row.Find(".element-name")

		elements.Each(func(_ int, el *goquery.Selection) {
			name := normalizeText(el.Text())
			if name == "" {
				return
			}

			attrSet := result[name]
			if attrSet == nil {
				attrSet = make(map[string]struct{})
				result[name] = attrSet
			}

			attributes.Each(func(_ int, attrSel *goquery.Selection) {
				attrName := normalizeText(attrSel.Text())
				if attrName == "" {
					return
				}
				attrSet[attrName] = struct{}{}
			})
		})
	})

	return result, nil
}

func ensureSymbolDefaults(m elementMap) {
	const symbolTag = "symbol"

	attrSet := m[symbolTag]
	if attrSet == nil {
		attrSet = make(map[string]struct{})
		m[symbolTag] = attrSet
	}
	attrSet["x"] = struct{}{}
	attrSet["y"] = struct{}{}
	attrSet["width"] = struct{}{}
	attrSet["height"] = struct{}{}
}

func fetchSVGTiny(ctx context.Context, client *http.Client) (elementMap, error) {
	doc, err := fetchDocument(ctx, client, svgTinyURL)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("#attributes .attribute")
	if rows.Length() == 0 {
		return nil, errors.New("could not find rows in SVG Tiny 1.2 index")
	}

	result := make(elementMap)
	rows.Each(func(_ int, row *goquery.Selection) {
		nameSel := row.Find(".attribute-name").First()
		attrName := normalizeText(nameSel.Text())
		if attrName == "" {
			return
		}

		row.Find(".element").Each(func(_ int, el *goquery.Selection) {
			elementName := normalizeText(el.Text())
			if elementName == "" {
				return
			}

			attrSet := result[elementName]
			if attrSet == nil {
				attrSet = make(map[string]struct{})
				result[elementName] = attrSet
			}
			attrSet[attrName] = struct{}{}
		})
	})

	return result, nil
}

func fetchSVG2(ctx context.Context, client *http.Client) (elementMap, error) {
	doc, err := fetchDocument(ctx, client, svg2URL)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tbody tr")
	if rows.Length() == 0 {
		return nil, errors.New("could not find rows in SVG 2 index")
	}

	result := make(elementMap)
	rows.Each(func(_ int, row *goquery.Selection) {
		nameSel := row.Find(".attr-name span").First()
		attrName := normalizeText(nameSel.Text())
		if attrName == "" {
			return
		}

		row.Find(".element-name span").Each(func(_ int, el *goquery.Selection) {
			elementName := normalizeText(el.Text())
			if elementName == "" {
				return
			}

			attrSet := result[elementName]
			if attrSet == nil {
				attrSet = make(map[string]struct{})
				result[elementName] = attrSet
			}
			attrSet[attrName] = struct{}{}
		})
	})

	return result, nil
}

func collectGlobals(maps []elementMap) map[string]struct{} {
	globals := make(map[string]struct{})

	for _, m := range maps {
		tagNames := keys(m)
		if len(tagNames) == 0 {
			continue
		}

		candidate := make(map[string]struct{})
		for _, attrSet := range m {
			for attr := range attrSet {
				if ignoreAttribute(attr) {
					continue
				}
				candidate[attr] = struct{}{}
			}
		}

		for attr := range candidate {
			isGlobal := true
			for _, tag := range tagNames {
				attrs := m[tag]
				if _, ok := attrs[attr]; !ok {
					isGlobal = false
					break
				}
			}
			if isGlobal {
				globals[attr] = struct{}{}
			}
		}
	}

	return globals
}

func mergeMaps(maps []elementMap, globals map[string]struct{}) map[string]map[string]struct{} {
	merged := make(map[string]map[string]struct{})
	for _, m := range maps {
		for element, attrSet := range m {
			target := merged[element]
			if target == nil {
				target = make(map[string]struct{})
				merged[element] = target
			}

			for attr := range attrSet {
				if _, skip := globals[attr]; skip {
					continue
				}
				if ignoreAttribute(attr) {
					continue
				}
				target[attr] = struct{}{}
			}
		}
	}
	return merged
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

func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "‘", "")
	value = strings.ReplaceAll(value, "’", "")
	return value
}

func ignoreAttribute(attribute string) bool {
	if attribute == "" {
		return true
	}

	if strings.HasPrefix(attribute, "aria-") {
		return true
	}

	if isEventHandler(attribute) {
		return true
	}

	if colon := strings.Index(attribute, ":"); colon != -1 {
		namespace := attribute[:colon]
		switch namespace {
		case "ev", "xlink", "xml":
			return true
		}
	}

	return false
}

func isEventHandler(attribute string) bool {
	if len(attribute) < 3 {
		return false
	}
	if !strings.HasPrefix(attribute, "on") {
		return false
	}
	r := rune(attribute[2])
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func keys(m elementMap) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}

func setToSortedSlice(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
