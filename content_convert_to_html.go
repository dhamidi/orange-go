package main

import (
	"bytes"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var MarkdownParser interface {
	Convert(string) (string, error)
}

type goldmarkParser struct {
	md goldmark.Markdown
}

func newGoldmarkParser() *goldmarkParser {
	return &goldmarkParser{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		),
	}
}

func (p *goldmarkParser) Convert(content string) (string, error) {
	var buf bytes.Buffer
	if err := p.md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func init() {
	MarkdownParser = newGoldmarkParser()
}

func ConvertContentToHTML(content string) string {
	asHTML, err := MarkdownParser.Convert(content)
	var root g.Node
	if err == nil {
		root = h.Article(g.Raw(asHTML))
	} else {
		root = h.Article(h.Pre(g.Text(content)), h.P(g.Textf("Error: %s", err)))
	}

	var buf bytes.Buffer
	root.Render(&buf)
	return buf.String()
}
