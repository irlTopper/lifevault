package utility

import (
	"fmt"
	"strings"
)

func BuildEmailString(name, email string) string {
	if name == email && strings.Contains(name, "@") {
		name = strings.Split(name, "@")[0]
	}

	return fmt.Sprintf("%s <%s>", name, email)
}
