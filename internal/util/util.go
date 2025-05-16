package util

import "strings"

// BuildHeaders converts a flat array of key-value pairs into a JSON object string.
// The input array should have keys at even indices and values at odd indices.
// This is used to create the header JSON string for HTMX requests in Gong components.
func BuildHeaders(headers []string) string {
	builder := &strings.Builder{}
	builder.WriteString("{")
	i := 0
	for i+1 < len(headers) {
		builder.WriteString(`"`)
		builder.WriteString(headers[i])
		builder.WriteString(`": "`)
		builder.WriteString(headers[i+1])
		builder.WriteString(`"`)
		if i < len(headers)-2 {
			builder.WriteString(", ")
		}
		i = i + 2
	}
	builder.WriteString("}")
	return builder.String()
}
