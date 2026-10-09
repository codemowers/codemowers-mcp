package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newMCPServer(sections []Advisory) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "codemowers-mcp", Version: version}, &mcp.ServerOptions{
		Instructions: "Engineering guidance focused on bare-metal Kubernetes and useful on managed cloud and local clusters. This service has no cluster access. Read the advisory index, developing/general for project decisions, and deploying/discovery and deploying/general before provisioning. Use the target cluster's advertised Driftmower tools or verified kubectl context for live capabilities. Follow deploying/storage for dependency ownership and native versus portable choices, and deploying/observability for Prometheus transport. Reuse recorded decisions. Resource examples do not establish availability; cluster-specific requirements govern the target cluster.",
	})
	var index strings.Builder
	index.WriteString("# Software engineering advisory\n\nPractical guidance focused on bare-metal Kubernetes, also useful on managed cloud and local clusters. This service has no cluster access. Start with `developing/general` for shared conventions and decision records; use `deploying/discovery` and `deploying/general` before provisioning. Frameworks are examples. Operator schemas and resource examples belong to the target cluster advisory.\n\n")
	for _, section := range sections {
		fmt.Fprintf(&index, "- [%s](%s) — %s\n", section.Path, section.URI, section.Title)
	}
	all := append([]Advisory{{Path: "index", URI: "engineering://advisory/index", Title: "Software engineering advisory", Markdown: index.String(), Source: "generated"}}, sections...)
	byPath := make(map[string]Advisory, len(all))
	for _, section := range all {
		byPath[section.Path] = section
		server.AddResource(&mcp.Resource{URI: section.URI, Name: section.Path, Title: section.Title, MIMEType: "text/markdown", Description: "General software engineering guidance: " + section.Title}, func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: section.URI, MIMEType: "text/markdown", Text: section.Markdown}}}, nil
		})
	}
	mcp.AddTool(server, &mcp.Tool{Name: "get_engineering_advisory", Title: "Read software engineering advisory", Description: "Read general guidance for developing, building and deploying applications. Omit path for the index. Operator examples belong to the target cluster's Driftmower advisory, when available. Read deploying/discovery for Driftmower or kubectl checks of CRDs, admission policies and storage classes, and deploying/general for configurable but unset Helm defaults. Also available as engineering://advisory/ resources.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: new(false), IdempotentHint: true, OpenWorldHint: new(false)}}, func(_ context.Context, _ *mcp.CallToolRequest, args struct {
		Path string `json:"path,omitempty" jsonschema:"Advisory path; omit for the index"`
	}) (*mcp.CallToolResult, Advisory, error) { if args.Path == "" {
		args.Path = "index"
	}; section, ok := byPath[args.Path]; if !ok {
		return nil, Advisory{}, fmt.Errorf("unknown advisory path: %s; read the index for available paths", args.Path)
	}; return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: section.Markdown}}}, section, nil })
	return server
}

func newHandler(sections []Advisory, publicURL string) (http.Handler, error) {
	origin, err := url.Parse(publicURL)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") {
		return nil, fmt.Errorf("PUBLIC_URL must be an HTTP(S) origin")
	}
	origin.Path = ""
	page, err := renderWebsite(sections, origin.String()+"/mcp")
	if err != nil {
		return nil, err
	}
	server := newMCPServer(sections)
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	mux := http.NewServeMux()
	mux.Handle("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Native MCP clients omit Origin. Browser requests must use the configured origin.
		if value := r.Header.Get("Origin"); value != "" && value != origin.String() {
			http.Error(w, "Origin is not allowed", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		transport.ServeHTTP(w, r)
	}))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; img-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", fmt.Sprint(len(page)))
		if r.Method != http.MethodHead {
			_, _ = w.Write(page)
		}
	})
	mux.HandleFunc("GET /guide", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusMovedPermanently) })
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) { http.ServeFileFS(w, r, assets, "web/style.css") })
	mux.HandleFunc("GET /codemowers-logo.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, assets, "web/codemowers-logo.png")
	})
	for _, section := range sections {
		mux.HandleFunc("GET /advisory/"+section.Path+".md", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			if r.Method != http.MethodHead {
				_, _ = fmt.Fprint(w, section.Markdown)
			}
		})
	}
	return mux, nil
}
