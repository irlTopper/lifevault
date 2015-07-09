package logger

import "regexp"

// Logger is an interface that log.Logger satisfies
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

var Log Logger
var TagMatchRegex = regexp.MustCompile(`\[(.*?)\]`)

func New(logger Logger) {
	Log = logger
}

func NewDefault() {
	Log = NewRevelLogger()
}

// Extract Tags from log messages.
// Tags are defined by any characters surrounded
// by brackets.
func ExtractTags(message string) (tags []string) {
	matches := TagMatchRegex.FindAllString(message, -1)
	for _, match := range matches {
		tags = append(tags, match[1:len(match)-1])
	}

	return tags
}
