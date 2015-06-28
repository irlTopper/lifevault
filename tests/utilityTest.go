package tests

import "github.com/revel/revel/testing"

// Must Embed `revel.TestSuite`
type MyAppTest struct {
	testing.TestSuite
}

// Run this before a request
func (t *MyAppTest) Before() {
	println("Set up")
}

// Run this after request
func (t *MyAppTest) After() {
	println("Tear down")
}

// Check main page is kinda there
func (t *MyAppTest) TestIndexPage() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html")
}

// Check if robots.txt exists
func (t *MyAppTest) TestRobotsPage() {
	t.Get("/robots.txt")
	t.AssertOk()
	t.AssertContentType("text/html")
}

// Will not appear in panel as it not start with `Test` case sensitive
func (t *MyAppTest) TEstFavIcon() {
	t.Get("/favicon.ico")
	t.AssertOk()
	t.AssertContentType("text/html")
}
