package main

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed content web
var assets embed.FS

type Advisory struct {
	Path     string        `json:"path"`
	URI      string        `json:"uri"`
	Title    string        `json:"title"`
	Markdown string        `json:"markdown"`
	Source   string        `json:"source"`
	HTML     template.HTML `json:"-"`
}

var advisoryPath = regexp.MustCompile(`^[a-z0-9-]+(?:\.[a-z0-9-]+)*(?:/[a-z0-9-]+(?:\.[a-z0-9-]+)*)*$`)

func loadAdvisories(source fs.FS) ([]Advisory, error) {
	var sections []Advisory
	err := fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		name := strings.TrimSuffix(path, ".md")
		if !advisoryPath.MatchString(name) || name == "index" {
			return fmt.Errorf("invalid advisory path: %s", path)
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		markdown := strings.ReplaceAll(string(data), "\r\n", "\n")
		heading, _, ok := strings.Cut(markdown, "\n")
		if !ok || !strings.HasPrefix(heading, "# ") || strings.TrimSpace(heading[2:]) == "" {
			return fmt.Errorf("advisory %s must start with a Markdown title", path)
		}
		sections = append(sections, Advisory{Path: name, URI: "engineering://advisory/" + name, Title: strings.TrimSpace(heading[2:]), Markdown: markdown, Source: "bundled"})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return nil, fmt.Errorf("no advisory Markdown found")
	}
	sort.Slice(sections, func(i, j int) bool { return sections[i].Path < sections[j].Path })
	return sections, nil
}

// Link document references written as inline code, without altering code blocks.
type referenceRenderer struct{ paths map[string]bool }

func (r referenceRenderer) RegisterFuncs(register renderer.NodeRendererFuncRegisterer) {
	register.Register(ast.KindCodeSpan, func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		value := string(node.Text(source))
		if r.paths[value] {
			_, err := fmt.Fprintf(w, `<a href="#%s"><code>%s</code></a>`, html.EscapeString(value), html.EscapeString(value))
			return ast.WalkSkipChildren, err
		}
		// Code span whitespace follows CommonMark's normalization rules.
		value = strings.ReplaceAll(value, "\n", " ")
		if strings.HasPrefix(value, " ") && strings.HasSuffix(value, " ") && strings.TrimSpace(value) != "" {
			value = value[1 : len(value)-1]
		}
		_, err := fmt.Fprintf(w, "<code>%s</code>", html.EscapeString(value))
		return ast.WalkSkipChildren, err
	})
}

type headingRenderer struct{}

func (headingRenderer) RegisterFuncs(register renderer.NodeRendererFuncRegisterer) {
	register.Register(ast.KindHeading, func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		heading := node.(*ast.Heading)
		value, _ := heading.AttributeString("id")
		id := html.EscapeString(string(value.([]byte)))
		if entering {
			_, err := fmt.Fprintf(w, `<h%d id="%s">`, heading.Level, id)
			return ast.WalkContinue, err
		}
		_, err := fmt.Fprintf(w, ` <a class="heading-permalink" href="#%s" aria-label="Link to this heading">#</a></h%d>`+"\n", id, heading.Level)
		return ast.WalkContinue, err
	})
}

type category struct {
	ID, Title string
	Sections  []Advisory
}

func renderWebsite(sections []Advisory, endpoint string) ([]byte, error) {
	paths := make(map[string]bool, len(sections))
	for _, section := range sections {
		paths[section.Path] = true
	}
	markdown := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()), goldmark.WithRendererOptions(renderer.WithNodeRenderers(util.Prioritized(referenceRenderer{paths}, 100), util.Prioritized(headingRenderer{}, 100))))
	groups := []category{{ID: "developing", Title: "Developing"}, {ID: "building", Title: "Building"}, {ID: "deploying", Title: "Deploying"}}
	for _, section := range sections {
		_, body, _ := strings.Cut(section.Markdown, "\n")
		source := []byte(body)
		document := markdown.Parser().Parse(text.NewReader(source))
		if err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if heading, ok := node.(*ast.Heading); ok && entering {
				heading.Level = min(heading.Level+2, 6)
				id, _ := heading.AttributeString("id")
				heading.SetAttributeString("id", []byte(section.Path+"/"+string(id.([]byte))))
			}
			if entering {
				var destination []byte
				switch link := node.(type) {
				case *ast.Link:
					destination = link.Destination
				case *ast.AutoLink:
					destination = link.URL(source)
				}
				if bytes.HasPrefix(destination, []byte("https://")) || bytes.HasPrefix(destination, []byte("http://")) {
					node.SetAttributeString("target", []byte("_blank"))
					node.SetAttributeString("rel", []byte("noopener noreferrer"))
				}
			}
			return ast.WalkContinue, nil
		}); err != nil {
			return nil, err
		}
		var rendered bytes.Buffer
		if err := markdown.Renderer().Render(&rendered, source, document); err != nil {
			return nil, err
		}
		// Goldmark omits raw HTML and unsafe URL schemes; the template escapes all other fields.
		section.HTML = template.HTML(rendered.String())
		group, _, _ := strings.Cut(section.Path, "/")
		index := -1
		for i := range groups {
			if groups[i].ID == group {
				index = i
				break
			}
		}
		if index < 0 {
			groups = append(groups, category{ID: group, Title: group})
			index = len(groups) - 1
		}
		groups[index].Sections = append(groups[index].Sections, section)
	}
	for i := range groups {
		sort.SliceStable(groups[i].Sections, func(a, b int) bool {
			return strings.HasSuffix(groups[i].Sections[a].Path, "/general") && !strings.HasSuffix(groups[i].Sections[b].Path, "/general")
		})
	}
	page, err := template.ParseFS(assets, "web/index.html")
	if err != nil {
		return nil, err
	}
	var result bytes.Buffer
	err = page.Execute(&result, struct {
		Categories []category
		Endpoint   string
		PublicURL  string
	}{groups, endpoint, strings.TrimSuffix(endpoint, "/mcp")})
	return result.Bytes(), err
}
