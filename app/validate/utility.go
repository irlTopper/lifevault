package validate

import "github.com/revel/revel"

func ValidateMatchIfSent(rc *revel.Controller, s string, r regexValidator) {
	// Check that it was sent
	if len(rc.Request.Form[s]) > 0 {
		rc.Validation.Match(rc.Params.Get(s), r.Expression).Message(r.Message, s)
	}
}

func ReturnErrors(errors []*revel.ValidationError) (e []string) {
	for _, err := range errors {
		e = append(e, err.Message)
	}
	return e
}
