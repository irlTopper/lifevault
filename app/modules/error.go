package modules

import (
	"strconv"
	"strings"

	"github.com/revel/revel"
)

type UserMessage struct {
	Status  int64
	Message string
}

// Checks the provided error (err) is a select one not found error
func SelectOneErrCheck(err error, message string) {
	if err != nil && err.Error() == "sql: no rows in result set" {
		panic(UserMessage{Status: 404, Message: message})
	}
}

// Checks the provided error (err). If an error exists, a message will be sent to all emails specified
// in config.json provided that developer mode is disabled. If it is enabled, the error will be outputted
// into the console instead.
func CheckErr(err error, message string, name string, errordata map[string]interface{}) {
	if err != nil && err.Error() != "sql: no rows in result set" {
		errordata["error"] = err
		errordata["name"] = name

		// Let's do some awesome shit to make the SQL reporting better
		if name == "SQL" {
			if SQL, ok := errordata["SQL"]; ok {
				errordata["OffendingSQL"] = SQL

				if strings.Contains(err.Error(), "at line") {
					line, err := strconv.ParseInt(err.Error()[strings.LastIndex(err.Error(), "at line")+8:], 10, 64)

					if err == nil {
						errordata["line"] = line
					}
				}

				if tmpParams, isok := errordata["Params"]; isok {
					if params, isok2 := tmpParams.([]interface{}); isok2 {
						var replacedSQL string = errordata["OffendingSQL"].(string)

						if len(params) == 1 {
							if paramMap, k := params[0].(map[string]interface{}); k {
								for key, value := range paramMap {
									if str, conv := value.(string); conv {
										replacedSQL = strings.Replace(replacedSQL, ":"+key, "'"+str+"'", -1)
									} else if val, conv := value.(int64); conv {
										replacedSQL = strings.Replace(replacedSQL, ":"+key, strconv.FormatInt(val, 10), -1)
									}
								}
							}
						} else if len(params) > 1 {
							for _, value := range params {
								if str, conv := value.(string); conv {
									replacedSQL = strings.Replace(replacedSQL, "?", "'"+str+"'", 1)
								} else if val, conv := value.(int64); conv {
									replacedSQL = strings.Replace(replacedSQL, "?", strconv.FormatInt(val, 10), 1)
								}
							}
						}

						errordata["OffendingSQL"] = replacedSQL
					}
				}
			}
		}

		revel.ERROR.Print(errordata)
		panic(errordata)
	}
}

// Returns a custom error message to the user in JSON format
func JSONError(rc *revel.Controller, httpStatusCode int, message string) revel.Result {
	rc.Response.Status = httpStatusCode
	rc.Response.Out.Header().Set("Status Code", strconv.Itoa(httpStatusCode)+" "+message)
	return rc.RenderJson(map[string]interface{}{"status": "error", "errors": []string{message}})
}

// Returns a JSON error message for validation errors
// The formatting could use some love.
func ShowValidationErrorsAsJSON(rc *revel.Controller) revel.Result {
	rc.Response.Status = 400
	rc.Response.Out.Header().Set("Status Code", strconv.Itoa(400)+" Error")
	return rc.RenderJson(map[string]interface{}{"status": "error", "errors": rc.Validation.ErrorMap()})
}
