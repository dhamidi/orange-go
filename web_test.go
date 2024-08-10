package main

import (
	"bytes"
	"log"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/html"
)

type WebTest struct {
	web  *WebApp
	logs *bytes.Buffer
	test *testing.T
}

func (w *WebTest) send(method, path string) *Response {
	req := httptest.NewRequest(method, path, nil)
	res := httptest.NewRecorder()
	w.web.ServeHTTP(res, req)
	return &Response{raw: res, test: w.test}
}

type Response struct {
	raw  *httptest.ResponseRecorder
	test *testing.T
}

func (res *Response) expectFormFields(fields ...*FormField) {
	doc, err := html.Parse(res.raw.Body)
	if err != nil {
		res.test.Fatalf("Error parsing response body as HTML: %v", err)
	}
	form := findNode(doc, "form")
	if form == nil {
		res.test.Fatalf("No form found in response")
	}
	for _, field := range fields {
		input := findNode(form, "input", "name", field.name)
		if input == nil {
			res.test.Fatalf("No input field with name %q found", field.name)
		}
		if getAttribute(input, "type") != field.kind {
			res.test.Fatalf("Field %q has type %q, expected %q", field.name, getAttribute(input, "type"), field.kind)
		}
	}
}

func findNode(n *html.Node, tag string, attrs ...string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		for i := 0; i < len(attrs); i += 2 {
			if getAttribute(n, attrs[i]) != attrs[i+1] {
				return nil
			}
		}
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNode(c, tag, attrs...); result != nil {
			return result
		}
	}
	return nil
}

func getAttribute(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

type FormField struct {
	name string
	kind string
}

func Field(name, kind string) *FormField {
	return &FormField{name: name, kind: kind}
}

func NewWebTest(t *testing.T) *WebTest {
	config := NewPlatformConfigForTest()
	app, _ := HackerNews(config)
	shell := NewDefaultShell(app)
	web := NewWebApp(app, shell)
	logs := bytes.NewBuffer(nil)
	web.logger = log.New(logs, "[web]", log.LstdFlags)
	return &WebTest{web: web, logs: logs, test: t}
}
