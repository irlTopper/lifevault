package modules

import (
	"math"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/guregu/null"
	"github.com/guregu/null/zero"
	"github.com/revel/revel"
)

// Go is fucking gay that I had to add this - Topper
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func CalculateMaxPages(count int64, pageSize int64) int64 {
	return int64(math.Ceil(float64(count) / float64(pageSize)))
}

// Sets a nullable valid string bla bla whatever - Sick of writing this shit
func SetNullableString(s *null.String, str string) {
	if str != "" {
		s.String = str
		s.Valid = true
	} else {
		s.Valid = false
	}
}

// Same as string but for integers
func SetNullableInt(i *null.Int, v string) {
	if v != "" {
		var err error

		i.Int64, err = strconv.ParseInt(v, 10, 64)

		if err == nil {
			i.Valid = true
		}
	} else {
		i.Valid = false
	}
}

func SetStringIfSet(rc *revel.Controller, s *string, str string) {
	if len(rc.Request.Form[str]) > 0 {
		*s = rc.Params.Get(str)
	}
}

func SetZeroStringIfSet(rc *revel.Controller, s *zero.String, str string) {
	if len(rc.Request.Form[str]) > 0 {
		s.String = rc.Params.Get(str)
		s.Valid = true
	}
}

func SetBoolIfSet(rc *revel.Controller, b *bool, str string) {
	if len(rc.Request.Form[str]) > 0 {
		val := false

		rc.Params.Bind(&val, str)

		*b = val
	}
}

func SetIntIfSet(rc *revel.Controller, i *int64, str string) {
	if len(rc.Request.Form[str]) > 0 {
		var val int64

		rc.Params.Bind(&val, str)

		*i = val
	}
}

// Oh man, goquery made this so easy
func HTMLtoPlainTextNoNewlines(html string) (output string) {
	q, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	return q.Text()
}

func Int64ArrayToString(ints *[]int64) (output string) {
	l := len(*ints)
	for i := 0; i < l; i++ {
		if i > 0 {
			output = output + ","
		}
		output = output + strconv.FormatInt((*ints)[i], 10)
	}
	return output
}
