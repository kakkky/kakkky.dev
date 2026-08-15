package view

func ParseMarkdownToHTML(src string) string {
	return src
}

type OutlineNode struct {
	Text     string
	Anchor   string
	Children []*OutlineNode
}

func ExtractOutlineFromMarkdown(src string) []*OutlineNode {
	return nil
}
