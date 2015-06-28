package filters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"time"

	"github.com/irlTopper/lifevault/app/modules"
	"github.com/kr/pretty"
	"github.com/revel/revel"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// PanicFilter wraps the action invocation in a protective defer blanket that
// converts panics into 500 error pages.
func PanicFilter(rc *revel.Controller, fc []revel.Filter) {
	defer func() {
		if err := recover(); err != nil {
			handleInvocationPanic(rc, err)
		}
	}()
	fc[0](rc, fc[1:])
}

// This function handles a panic in an action invocation.
// It cleans up the stack trace, logs it, and displays an error page.
func handleInvocationPanic(rc *revel.Controller, err interface{}) {
	if userMsg, ok := err.(modules.UserMessage); ok {
		rc.Response.Status = int(userMsg.Status)
		rc.Result = rc.RenderText(userMsg.Message)
		return
	} else if strError, _ := err.(string); strError == "Duplicate" {
		rc.Response.Status = 400
		rc.Result = rc.RenderJson(map[string]interface{}{"errors": []string{"An object with those details already exists"}})
		return
	}

	runtimeOffset := 7
	stackOffset := 8

	errorSource := ""
	errorLine := ""

	renderErrors := map[string]interface{}{
		"request": pretty.Sprintf("%# v", rc.Request.Request),
	}

	if data, ok := err.(map[string]interface{}); ok {
		err = data["error"]
		runtimeOffset = 7
		stackOffset = 8

		if SQL, ok := data["OffendingSQL"]; ok {
			renderErrors["SQL"] = SQL
		}
	}

	// Let's figure out where this error came from specifically
	if _, filename, line, ok := runtime.Caller(runtimeOffset); ok {
		errorSource = filename
		errorLine = strconv.Itoa(line)
	}

	renderErrors["stack"] = string(stack(stackOffset))
	renderErrors["fullStack"] = string(stack(0))
	renderErrors["errorSource"] = errorSource
	renderErrors["errorLine"] = errorLine
	renderErrors["title"] = err

	rc.RenderArgs = renderErrors
	rc.Response.Status = 500

	if revel.DevMode {
		rc.Result = rc.RenderTemplate("errors/panic.html")
	} else {
		title, _ := renderErrors["title"].(string)

		devEmails := map[string]string{
			"Robert O'Leary":  "teamworkrobert@gmail.com",
			"Peter Coppinger": "peter@teamwork.com",
			"Brandon Hansen":  "contact+tmpanic@aisforarray.com",
		}

		// TODO SEND EMAIL
		modules.SendEmailTemplate(&modules.EmailFields{
			Subject:      "PANIC:" + title,
			From:         "panic@teamworkdesk.com",
			Name:         "Teamwork Desk Panic",
			To:           devEmails,
			TemplateFile: "errors/panic.html",
			Data:         renderErrors,
		})

		rc.Result = rc.RenderJson(map[string]interface{}{
			"status": "Internal Error",
			"error":  "An internal error occurred, sorry about that - the team have been notified of the issue!",
		})

		data, _ := json.Marshal(renderErrors)

		modules.DB.Exec(rc, "INSERT INTO paniclogs (data, createdAt) VALUES (?, ?)", data, time.Now().UTC())
	}

	return
}

func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the Pc.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
