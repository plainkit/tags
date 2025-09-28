# tags

This library exists because a spec-accurate catalogue of HTML, SVG, and MathML elements is essential for the tooling that depends on it. Hand-maintained lists drift, so the project scrapes the standards, normalizes them, and commits the result as code that is ready for import.

## Installation

```bash
go get github.com/plainkit/tags
```

## Package overview

The root package exposes three indexes:

- `tags.HTML` lists every HTML element, the attributes it accepts, and which of those attributes are boolean.
- `tags.SVG` lists every SVG element together with its attribute set and the collection of SVG global attributes.
- `tags.MathML` is the set of MathML element names.

Each index includes metadata with the upstream sources and the schema version used during generation. You can compare the schema version to detect when the published data changes.

### HTML example

```go
package main

import (
    "fmt"

    "github.com/plainkit/tags"
)

func main() {
    button := tags.HTML.Elements["button"]
    fmt.Printf("button is empty? %v\n", button.Empty)

    for _, attr := range button.Attributes {
        if attr.Boolean {
            fmt.Printf("%s (boolean)\n", attr.Name)
            continue
        }
        fmt.Println(attr.Name)
    }

    fmt.Println("Global attribute count:", len(tags.HTML.Globals))
}
```

### SVG example

```go
rect := tags.SVG.Elements["rect"]
fmt.Println("SVG rect attributes:", rect.Attributes)
fmt.Println("Globals:", tags.SVG.Globals)
```

### MathML example

```go
for _, name := range tags.MathML.Tags {
    // Build an allow-list or completion list.
    _ = name
}
```

## JSON exports

If you are not writing Go, you can consume the generated JSON indexes under `./json/`. They mirror the Go structures, so you get HTML element definitions, SVG element definitions, and the MathML tag list in a language-agnostic format.

## Regenerating the data

The checked-in files are built by the standalone parser under `./parser`. Regenerating requires network access because it downloads the WHATWG HTML specification and the W3C SVG and MathML specifications. Run:

```bash
cd parser && go run .
```

That command refreshes the JSON payloads and the Go source files in the repository root. Regeneration is deterministic; if the upstream specs have not changed, you will get bit-for-bit identical output.

## Notes

- The data set intentionally excludes custom elements and event handler attributes so the indexes stay focused on markup defined by the specifications.
- Boolean attribute information is preserved for HTML so you can emit the compact forms (`disabled` instead of `disabled="disabled"`).
- SVG globals are deduplicated across the SVG 1.1, Tiny 1.2, and SVG 2 sources, with a best-effort normalization of attribute names.
