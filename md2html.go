package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	//go:embed template.html
	htmlTemplate []byte
)

type MdOpts struct {
	Preview    bool
	Standalone bool
	Mathjax    bool
}

func NewMdOpts(opts ...func(*MdOpts)) *MdOpts {
	m := &MdOpts{}
	m.Preview = false
	m.Standalone = false
	m.Mathjax = false
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func EnablePreviewMode() func(*MdOpts) {
	return func(m *MdOpts) {
		m.Preview = true
	}
}

func StandaloneDocument() func(*MdOpts) {
	return func(m *MdOpts) {
		m.Standalone = true
	}
}

func EnableMathjax() func(*MdOpts) {
	return func(m *MdOpts) {
		m.Mathjax = true
	}
}

func ProcessHeadings(doc ast.Node, renderer *html.Renderer) (string, string) {
	var firstHeading bool = true
	var docTitle string = ""
	var curLevels []int = []int{0, 0, 0} // this refers to h2, h3, h4 counters (anything higher than h4 is not ignored)
	var toc string = "<ul style=\"list-style-type:none;\">"
	var headingCounter int = 0

	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering {
			if firstHeading {
				firstHeading = false
				titleNode := ast.Document{}
				titleNode.SetChildren(node.GetChildren())
				ast.WalkFunc(&titleNode, func(node ast.Node, entering bool) ast.WalkStatus {
					if text, ok := node.(*ast.Text); ok && entering {
						docTitle += string(text.Literal)
					}
					return ast.GoToNext
				})
				ast.RemoveFromTree(heading)
			}

			if heading.Level >= 2 && heading.Level <= 4 {
				headingCounter++
				heading.HeadingID = fmt.Sprintf("heading-%d", headingCounter)
				index := heading.Level - 2

				tmp := ast.Document{}
				tmp.SetChildren(heading.GetChildren())
				html := markdown.Render(&tmp, renderer)
				tocEntry := fmt.Sprintf("<a href=\"#%s\">%s</a>", heading.HeadingID, html)

				curLevels[index]++
				for i := index + 1; i < len(curLevels); i++ {
					if curLevels[i] != 0 {
						curLevels[i] = 0
					}
				}

				var sectionNumber string
				for i := 0; i <= index; i++ {
					sectionNumber += strconv.Itoa(curLevels[i]) + "."
				}
				sectionNumber += " "

				indent := fmt.Sprintf("%dpx", 16*(index-2))
				toc += "<li style=\"margin-left: " + indent + "\">" + sectionNumber + tocEntry + "</li>"

				sectionNumberNode := &ast.Text{Leaf: ast.Leaf{Literal: []byte(sectionNumber)}}
				heading.SetChildren(append([]ast.Node{sectionNumberNode}, heading.GetChildren()...))
			}
		}
		return ast.GoToNext
	})

	toc += "</ul>"
	return docTitle, toc
}

func PublishTimeString() string {
	now := time.Now()
	day := now.Format("January 2, 2006")
	time := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	return "Published on " + day + " at " + time
}

func ConvertMarkdownToHtml(md []byte, options *MdOpts) (string, []byte) {
	hash := sha256.Sum256(md)
	id := hex.EncodeToString(hash[:])[:16]

	parser := parser.NewWithExtensions(parser.CommonExtensions)
	opts := html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank | html.SkipHTML}
	renderer := html.NewRenderer(opts)

	doc := parser.Parse(md)
	title, toc := ProcessHeadings(doc, renderer)

	maybeUnsafeHTML := markdown.Render(doc, renderer)
	policy := bluemonday.UGCPolicy()
	content := string(policy.SanitizeBytes(maybeUnsafeHTML))

	tmpl, err := template.New("foo").Parse(string(htmlTemplate))
	if err != nil {
		log.Fatal("Failed to parse static HTML template with error", err)
	}

	params := map[string]any{
		"Title":           title,
		"Content":         template.HTML(content),
		"TableOfContents": template.HTML(toc),
		"Date":            PublishTimeString(),
		"Mathjax":         options.Mathjax,
		"Standalone":      options.Standalone,
		"Preview":         options.Preview,
	}

	var out bytes.Buffer
	tmpl.Execute(&out, params)
	return id, out.Bytes()
}
