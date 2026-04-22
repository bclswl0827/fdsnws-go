package utils

import (
	"encoding/xml"
	"sort"
	"strings"
)

// BuildWADL creates the common WADL shape used by built-in components.
func BuildWADL(base string, parameters []string, queryPost bool, extra map[string][]string) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<application xmlns="http://wadl.dev.java.net/2009/02"><resources base="`)
	_ = xml.EscapeText(&builder, []byte(base))
	builder.WriteString(`"><resource path="query"><method name="GET"><request>`)
	for _, parameter := range parameters {
		builder.WriteString(`<param style="query" name="`)
		_ = xml.EscapeText(&builder, []byte(parameter))
		builder.WriteString(`" type="xs:string" xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`)
	}
	builder.WriteString(`</request><response status="200"/></method>`)
	if queryPost {
		builder.WriteString(`<method name="POST"/>`)
	}
	builder.WriteString(`</resource>`)
	paths := make([]string, 0, len(extra))
	for path := range extra {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		builder.WriteString(`<resource path="`)
		_ = xml.EscapeText(&builder, []byte(path))
		builder.WriteString(`">`)
		for _, method := range extra[path] {
			builder.WriteString(`<method name="`)
			_ = xml.EscapeText(&builder, []byte(method))
			builder.WriteString(`"/>`)
		}
		builder.WriteString(`</resource>`)
	}
	builder.WriteString(`<resource path="version"><method name="GET"/></resource><resource path="application.wadl"><method name="GET"/></resource></resources></application>`)
	return builder.String()
}
