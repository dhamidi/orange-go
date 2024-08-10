package main

import "testing"

func TestWebApp_PageLogin_hasLogin_Form(t *testing.T) {
	w := NewWebTest(t)
	res := w.send("GET", "/login")
	res.expectFormFields(Field("username", "text"), Field("password", "password"))
}
