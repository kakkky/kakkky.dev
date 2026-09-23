package js

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/a-h/templ"
)

const importMapTemplateHTML = `<script type="importmap">%s</script>`

var CodeMirrorImports = map[string]string{
	"codemirror":                "https://esm.sh/codemirror@6.0.1",
	"@codemirror/lang-markdown": "https://esm.sh/@codemirror/lang-markdown@6.3.1",
	"@codemirror/language":      "https://esm.sh/@codemirror/language@^6.0.0",
	"@lezer/highlight":          "https://esm.sh/@lezer/highlight@^1.0.0",
}

type importMap map[string]string

func (m importMap) Render(_ context.Context, w io.Writer) error {
	body, err := json.Marshal(struct {
		Imports map[string]string `json:"imports"`
	}{Imports: m})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, importMapTemplateHTML, body)
	return err
}

// ImportMap は与えられた複数のエントリセットをマージして <script type="importmap"> を返す。
func ImportMap(entrySets ...map[string]string) templ.Component {
	merged := importMap{}
	for _, s := range entrySets {
		for k, v := range s {
			merged[k] = v
		}
	}
	return merged
}
