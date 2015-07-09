package logger

import (
	"fmt"
	"os"

	"github.com/revel/revel"
)

type RevelLogger struct {
	Prefix string
}

func NewRevelLogger() Logger {
	return new(RevelLogger)
}

func (log RevelLogger) Print(v ...interface{}) {
	revel.INFO.Print(v...)
}

func (log RevelLogger) Printf(format string, v ...interface{}) {
	revel.INFO.Printf(format, v...)
}

func (log RevelLogger) Println(v ...interface{}) {
	revel.INFO.Println(v...)
}

func (log RevelLogger) Fatal(v ...interface{}) {
	revel.ERROR.Print(v...)
	os.Exit(1)
}

func (log RevelLogger) Fatalf(format string, v ...interface{}) {
	revel.ERROR.Printf(format, v...)
	os.Exit(1)
}

func (log RevelLogger) Fatalln(v ...interface{}) {
	revel.ERROR.Println(v...)
	os.Exit(1)
}

// Conform to the interface here, but don't panic as most app
// code is not expecting that at this point
func (log RevelLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	revel.ERROR.Print(s)
}

func (log RevelLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	revel.ERROR.Print(s)
}

func (log RevelLogger) Panicln(v ...interface{}) {
	s := fmt.Sprint(v...)
	revel.ERROR.Println(s)
}
