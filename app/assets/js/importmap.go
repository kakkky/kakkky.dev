package js

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

const codeMirrorImportMapHTML = `<script type="importmap">
{
  "imports": {
    "codemirror": "https://esm.sh/codemirror@6.0.1",
    "@codemirror/lang-markdown": "https://esm.sh/@codemirror/lang-markdown@6.3.1",
    "@codemirror/language": "https://esm.sh/@codemirror/language@^6.0.0",
    "@lezer/highlight": "https://esm.sh/@lezer/highlight@^1.0.0"
  }
}
</script>`

type importMap string

func (m importMap) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(m))
	return err
}

func CodeMirrorImportMap() templ.Component {
	return importMap(codeMirrorImportMapHTML)
}
