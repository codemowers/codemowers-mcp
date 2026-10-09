package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

func bundled(t *testing.T) []Advisory {
	t.Helper()
	content, err := fs.Sub(assets, "content")
	if err != nil {
		t.Fatal(err)
	}
	sections, err := loadAdvisories(content)
	if err != nil {
		t.Fatal(err)
	}
	return sections
}

func callRPC(t *testing.T, handler http.Handler, method string, params any) map[string]any {
	t.Helper()
	data, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": "test-request", "method": method, "params": params})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "https://mcp.codemowers.io/mcp", strings.NewReader(string(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-06-18")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("%s: HTTP %d: %s", method, response.Code, response.Body.String())
	}
	if response.Header().Get("Mcp-Session-Id") != "" || response.Header().Get("WWW-Authenticate") != "" {
		t.Fatal("public stateless service returned session or auth headers")
	}
	var envelope map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["id"] != "test-request" || envelope["jsonrpc"] != "2.0" {
		t.Fatalf("invalid envelope: %v", envelope)
	}
	return envelope
}

func TestPublicAdvisoryProtocol(t *testing.T) {
	sections := bundled(t)
	handler, err := newHandler(sections, "https://mcp.codemowers.io")
	if err != nil {
		t.Fatal(err)
	}
	initialized := callRPC(t, handler, "initialize", map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "integration-test", "version": "1"}})
	info := initialized["result"].(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != "codemowers-mcp" {
		t.Fatalf("wrong server identity: %v", info)
	}
	tools := callRPC(t, handler, "tools/list", map[string]any{})["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 1 || tools[0].(map[string]any)["name"] != "get_engineering_advisory" {
		t.Fatalf("unexpected tools: %v", tools)
	}
	resources := callRPC(t, handler, "resources/list", map[string]any{})["result"].(map[string]any)["resources"].([]any)
	if len(resources) != len(sections)+1 {
		t.Fatalf("got %d resources for %d documents", len(resources), len(sections))
	}
	index := callRPC(t, handler, "tools/call", map[string]any{"name": "get_engineering_advisory", "arguments": map[string]any{}})["result"].(map[string]any)["structuredContent"].(map[string]any)
	if index["path"] != "index" {
		t.Fatalf("wrong default: %v", index)
	}
	for _, section := range sections {
		t.Run(section.Path, func(t *testing.T) {
			tool := callRPC(t, handler, "tools/call", map[string]any{"name": "get_engineering_advisory", "arguments": map[string]any{"path": section.Path}})["result"].(map[string]any)
			structured := tool["structuredContent"].(map[string]any)
			if structured["markdown"] != section.Markdown || structured["cluster"] != nil {
				t.Fatal("tool content differs or is incorrectly cluster-scoped")
			}
			resource := callRPC(t, handler, "resources/read", map[string]any{"uri": section.URI})["result"].(map[string]any)["contents"].([]any)[0].(map[string]any)
			if resource["text"] != section.Markdown {
				t.Fatal("resource content differs")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/advisory/"+section.Path+".md", nil))
			if response.Code != 200 || response.Body.String() != section.Markdown {
				t.Fatal("raw Markdown content differs")
			}
		})
	}
	invalid := callRPC(t, handler, "tools/call", map[string]any{"name": "get_engineering_advisory", "arguments": map[string]any{"path": "../../etc/passwd"}})["result"].(map[string]any)
	if invalid["isError"] != true {
		t.Fatal("unknown tool path was accepted")
	}
	if callRPC(t, handler, "resources/read", map[string]any{"uri": "engineering://advisory/missing"})["error"] == nil {
		t.Fatal("unknown resource was accepted")
	}
}

func TestWebsiteAndHTTPBoundaries(t *testing.T) {
	handler, err := newHandler(bundled(t), "https://mcp.codemowers.io")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path string
		code         int
	}{{"GET", "/", 200}, {"HEAD", "/", 200}, {"GET", "/style.css", 200}, {"GET", "/guide", 301}, {"GET", "/healthz", 404}, {"GET", "/readyz", 404}, {"GET", "/inspector.js", 404}, {"GET", "/.well-known/oauth-protected-resource", 404}, {"GET", "/mcp", 405}, {"DELETE", "/mcp", 405}} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
		if rr.Code != tc.code {
			t.Errorf("%s %s: got %d, want %d", tc.method, tc.path, rr.Code, tc.code)
		}
		if tc.method == "HEAD" && rr.Body.Len() != 0 {
			t.Error("HEAD returned body")
		}
		if tc.path == "/" && tc.method == "GET" {
			body := rr.Body.String()
			for _, want := range []string{"Zen of", "Kubernetes.", "https://mcp.codemowers.io/mcp", `id="developing/general"`, `<table>`, `href="#deploying/networking"`} {
				if !strings.Contains(body, want) {
					t.Errorf("homepage missing %s", want)
				}
			}
			if strings.Contains(body, "<script") || strings.Contains(body, "MCP Explorer") {
				t.Error("unexpected browser application")
			}
		}
	}
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{}`))
	req.Header.Set("Origin", "https://unrelated.example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatal("foreign browser origin accepted")
	}
	for _, value := range []string{"ftp://example.com", "https://example.com/path", "https://user:pass@example.com", "https://example.com?x=y"} {
		if _, err := newHandler(bundled(t), value); err == nil {
			t.Errorf("accepted invalid origin %q", value)
		}
	}
}

func TestPreviewLogo(t *testing.T) {
	handler, err := newHandler(bundled(t), "https://mcp.codemowers.io")
	if err != nil {
		t.Fatal(err)
	}
	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(page.Body.String(), `<meta property="og:image" content="https://mcp.codemowers.io/codemowers-logo.png">`) {
		t.Fatal("preview image URL missing")
	}
	expected, err := assets.ReadFile("web/codemowers-logo.png")
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "HEAD"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/codemowers-logo.png", nil)
		request.Header.Set("User-Agent", "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/png" {
			t.Fatalf("%s preview image: HTTP %d, type %q", method, response.Code, response.Header().Get("Content-Type"))
		}
		if method == "HEAD" {
			if response.Body.Len() != 0 {
				t.Fatal("HEAD returned image body")
			}
			continue
		}
		if !bytes.Equal(response.Body.Bytes(), expected) {
			t.Fatal("preview image differs from bundled logo")
		}
		config, err := png.DecodeConfig(response.Body)
		if err != nil || config.Width != 868 || config.Height != 868 {
			t.Fatalf("preview image dimensions differ from metadata: %+v, %v", config, err)
		}
	}
}

func TestHeadingPermalinksAndGitHubSources(t *testing.T) {
	sections, err := loadAdvisories(fstest.MapFS{
		"developing/first.md": &fstest.MapFile{Data: []byte("# First\n\n## Shared title\n\n## Shared title\n")},
		"building/second.md":  &fstest.MapFile{Data: []byte("# Second\n\n## Shared title\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := renderWebsite(append(bundled(t), sections...), "https://mcp.codemowers.io/mcp")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, match := range regexp.MustCompile(`\bid="([^"]+)"`).FindAllSubmatch(page, -1) {
		id := string(match[1])
		if ids[id] {
			t.Errorf("duplicate anchor %q", id)
		}
		ids[id] = true
	}
	for _, match := range regexp.MustCompile(`href="#([^"]+)"`).FindAllSubmatch(page, -1) {
		target, err := url.PathUnescape(string(match[1]))
		if err != nil {
			t.Fatal(err)
		}
		if !ids[target] {
			t.Errorf("permalink has no target: %s", match[1])
		}
	}
	for _, section := range sections {
		for _, want := range []string{
			`class="heading-link" href="#` + strings.ReplaceAll(section.Path, "/", "%2f") + `"`,
			`href="#` + section.Path + `/shared-title"`,
			`href="https://github.com/codemowers/codemowers-mcp/blob/main/content/` + section.Path + `.md"`,
			`aria-label="View ` + section.Title + ` on GitHub"`,
		} {
			if !strings.Contains(string(page), want) {
				t.Errorf("missing heading or source link: %s", want)
			}
		}
	}
	if len(regexp.MustCompile(`href="#(?:developing/first|building/second)/shared-title[^"]*"`).FindAll(page, -1)) != 3 {
		t.Fatal("not every subsection has a permalink")
	}
}

func TestContentValidationAndSafeRendering(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"index.md", "# Reserved\n"}, {"bad..path.md", "# Invalid\n"}, {"doc.md", "Missing title\n"}} {
		_, err := loadAdvisories(fstest.MapFS{tc.name: &fstest.MapFile{Data: []byte(tc.body)}})
		if err == nil {
			t.Errorf("accepted %s: %s", tc.name, tc.body)
		}
	}
	sections, err := loadAdvisories(fstest.MapFS{"example/doc.md": &fstest.MapFile{Data: []byte("# Example\n\n<script>alert(1)</script>\n\n[bad](javascript:alert)\n\n`<unsafe>`\n\n## Detail\n")}})
	if err != nil {
		t.Fatal(err)
	}
	page, err := renderWebsite(sections, "https://mcp.codemowers.io/mcp")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"<script>", `href="javascript:`, "<unsafe>"} {
		if strings.Contains(string(page), bad) {
			t.Errorf("unsafe output %q", bad)
		}
	}
	if !strings.Contains(string(page), `<h4 id="example/doc/detail">Detail <a class="heading-permalink" href="#example/doc/detail"`) {
		t.Fatal("document heading hierarchy was not preserved")
	}
	// Catch dangling cross-document references after extraction.
	known := map[string]bool{}
	for _, section := range bundled(t) {
		known[section.Path] = true
	}
	for _, section := range bundled(t) {
		for _, token := range strings.Split(section.Markdown, "`") {
			if strings.Contains(token, "\n") || strings.Contains(token, " ") {
				continue
			}
			if (strings.HasPrefix(token, "crds/") || strings.HasPrefix(token, "developing/") || strings.HasPrefix(token, "building/") || strings.HasPrefix(token, "deploying/")) && !known[token] {
				t.Error(fmt.Sprintf("%s references missing %s", section.Path, token))
			}
		}
	}
}
