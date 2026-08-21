package main

import (
	"log"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	style := styles.Get("base16-snazzy")
	formatter := html.New(html.WithClasses(true))
	if err := formatter.WriteCSS(os.Stdout, style); err != nil {
		log.Fatal(err)
	}
}
