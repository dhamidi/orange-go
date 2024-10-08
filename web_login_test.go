package main

import "testing"

func TestWebApp_PageLogin_hasLogin_Form(t *testing.T) {
	w := NewWebTest(t)
	res := w.send("GET", "/login")
	res.expectFormFields(Field("username", "text"), Field("password", "password"))
}

func TestWebApp_PageItem_displays_submissions(t *testing.T) {
	w := NewWebTest(t)
	w.RegisterUser("guest")
	session := w.LogInAs("guest")
	session.SubmitItem("https://example.com", "Example")
	session.ExpectPath("/")
	session.GoToSubmission(0)
}
