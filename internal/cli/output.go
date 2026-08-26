package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// writeJSON encodes value v as indented JSON to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// writeText writes a string followed by a newline to w.
func writeText(w io.Writer, s string) error {
	_, err := fmt.Fprintln(w, s)
	return err
}
