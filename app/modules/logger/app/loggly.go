package logger

import (
	"fmt"
	"os"

	"github.com/revel/revel"
	"github.com/segmentio/go-loggly"
)

// LogglyLogger implements the logging interface.  The actual sending
// takes place in a separate, buffered thread and should not affect performance
// in a tangible manner
type LogglyLogger struct {
	Prefix string
	client *loggly.Client
}

func NewLogglyLogger() Logger {
	token := revel.Config.StringDefault("loggly.token", "579f7bd1-e16a-406a-a68b-aea30884ca78")
	return LogglyLogger{
		client: loggly.New(token, "desk", revel.Config.StringDefault("loggly.servername", "dev")),
	}
}

func (log LogglyLogger) Print(v ...interface{}) {
	log.client.Info(fmt.Sprint(v...))
}

func (log LogglyLogger) Printf(format string, v ...interface{}) {
	log.client.Info(fmt.Sprintf(format, v...))
}

func (log LogglyLogger) Println(v ...interface{}) {
	log.client.Info(fmt.Sprintln(v...))
}

func (log LogglyLogger) Fatal(v ...interface{}) {
	log.client.Emergency(fmt.Sprint(v...))
	os.Exit(1)
}

func (log LogglyLogger) Fatalf(format string, v ...interface{}) {
	log.client.Emergency(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (log LogglyLogger) Fatalln(v ...interface{}) {
	log.client.Emergency(fmt.Sprintln(v...))
	os.Exit(1)
}

// Conform to the interface here, but don't panic as most app
// code is not expecting that at this point
func (log LogglyLogger) Panic(v ...interface{}) {
	log.client.Error(fmt.Sprint(v...))
}

func (log LogglyLogger) Panicf(format string, v ...interface{}) {
	log.client.Error(fmt.Sprintf(format, v...))
}

func (log LogglyLogger) Panicln(v ...interface{}) {
	log.client.Error(fmt.Sprintln(v...))
}
