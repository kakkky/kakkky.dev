package js

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

const echartsScriptHTML = `<script src="https://cdn.jsdelivr.net/npm/echarts@5.5.1/dist/echarts.min.js" defer></script>`

type echartsScript string

func (s echartsScript) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(s))
	return err
}

func EChartsScript() templ.Component {
	return echartsScript(echartsScriptHTML)
}
