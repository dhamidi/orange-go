package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/html"
)

type RequestOption interface {
	BuildRequest(req *http.Request)
}

type RequestOptionFunc func(req *http.Request)

func (f RequestOptionFunc) BuildRequest(req *http.Request) { f(req) }

func SetCookie(name, value string) RequestOption {
	return RequestOptionFunc(func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	})
}

type WebTest struct {
	web  *WebApp
	logs *bytes.Buffer
	test *testing.T
}

func (w *WebTest) send(method, path string) *Response {
	w.test.Logf("-> %s %s", method, path)
	req := httptest.NewRequest(method, path, nil)
	res := httptest.NewRecorder()
	w.web.ServeHTTP(res, req)
	w.test.Logf("<- %s %s %d", method, path, res.Result().StatusCode)
	return &Response{raw: res, test: w.test}
}

func (w *WebTest) post(path string, body url.Values, options ...RequestOption) *Response {
	w.test.Logf("-> POST %s %s", path, body)
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, opt := range options {
		opt.BuildRequest(req)
	}
	res := httptest.NewRecorder()
	w.web.ServeHTTP(res, req)
	w.test.Logf("<- POST %s %d %s", path, res.Result().StatusCode, res.Result().Header.Get("Location"))
	return &Response{raw: res, test: w.test}
}

type Response struct {
	raw  *httptest.ResponseRecorder
	test *testing.T
}

func (res *Response) Location() *url.URL {
	loc, err := res.raw.Result().Location()
	if err != nil {
		if errors.Is(err, http.ErrNoLocation) {
			return nil
		}
		res.test.Fatalf("Failed to parse location: %s", err)
	}
	return loc
}

func (res *Response) Find(tag string, attrs ...string) *html.Node {
	doc, err := html.Parse(res.raw.Body)
	if err != nil {
		res.test.Fatalf("Error parsing response body as HTML: %v", err)
	}
	return findNode(doc, tag, attrs...)
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

type WebSession struct {
	test      *WebTest
	sessionID string
	location  *url.URL
}

func (s *WebSession) ExpectPath(p string) {
	if s.location.Path != p {
		s.test.test.Fatalf("Expected path %q, current location is %q", p, s.location.Path)
	}
}

func (s *WebSession) SubmitItem(u, title string) {
	response := s.test.post("/submit", url.Values{
		"title": []string{title},
		"url":   []string{u},
	}, SetCookie("session_id", s.sessionID))
	s.location = response.Location()
}

func (w *WebTest) RegisterUser(username string) {
	password, err := HashPassword("password")
	if err != nil {
		w.test.Fatalf("failed to hash password for %s: %s", username, err)
	}
	err = w.web.app.HandleCommand(&SignUpUser{
		PasswordHash: *password,
		Username:     username,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		w.test.Fatalf("failed to sign up user %s: %s", username, err)
	}
}

func (w *WebTest) LogInAs(username string) *WebSession {
	response := w.post("/login", url.Values{"username": []string{username}, "password": []string{"password"}})
	var sessionID *string = nil
	for _, cookie := range response.raw.Result().Cookies() {
		if cookie.Name == "session_id" {
			sessionID = &cookie.Value
			break
		}
	}
	if sessionID == nil {
		w.test.Fatalf("No sessionID cookie returned: %s", response.raw.Body.String())
	}
	return &WebSession{test: w, sessionID: *sessionID, location: response.Location()}
}

func (s *WebSession) GoToSubmission(index int) {
	q := NewFrontpageQuery(nil)
	if err := s.test.web.app.HandleQuery(q); err != nil {
		s.test.test.Fatalf("Failed to query submissions: %s", err)
	}
	if len(q.Submissions) < index {
		s.test.test.Fatalf("Index out of range (%d submissions found, looking for %d)", len(q.Submissions), index)
	}
	itemURL := &url.URL{Path: "/item", RawQuery: "id=" + q.Submissions[index].ItemID}
	response := s.test.send("GET", itemURL.String())
	if loc := response.Location(); loc != nil {
		s.location = response.Location()
	}
}
