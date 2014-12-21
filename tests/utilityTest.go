package tests

import (
	"github.com/irlTopper/ohlife2/app/modules"
	"github.com/revel/revel"
)

type UtilityTest struct {
	revel.TestSuite
}

type HtmlTest struct {
	Expected, Input string
}

func (t *UtilityTest) Test_StripsHtmlChars() {
	tests := []HtmlTest{
		{
			Expected: "This is a test",
			Input:    "<p>This is a test</p>",
		},
		{
			Expected: "This is a test",
			Input:    "<p>This <p>is a test</p>",
		},
		{
			Expected: "This is a test",
			Input:    "<tr>This is a test</tr>",
		},
		{
			Expected: "This is a test",
			Input:    "<tr>This is a <br>test</tr>",
		},
		{
			Expected: "This is a test",
			Input:    "<tr>This is a <br />test</tr>",
		},
	}

	for _, test := range tests {
		t.AssertEqual(test.Expected, modules.HTMLtoPlainTextNoNewlines(test.Input))
	}
}
