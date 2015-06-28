package interceptors

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/irlTopper/lifevault/app/models"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
)

type Authentication struct {
	*revel.Controller
	User              *models.Session // Very annoying that we can't use "Session" - taken by revel cookie
	RequestLog        *modules.RequestLog
	startTimeUnixNano int64
}

func (c *Authentication) DecodeUserSession() revel.Result {
	var session models.Session

	// BASIC AUTH
	// If a new header has been sent up, we must perform authenication
	// Need to make sure they're not already logged in!
	if auth := c.Request.Header.Get("Authorization"); auth != "" {
		host := c.Request.Host

		if strings.Contains(host, ":") {
			host = host[:strings.LastIndex(host, ":")]
		}

		session, err := models.AuthenticateUser(auth, true, host, c.Controller)
		if err != nil {
			c.Response.Status = 401
			c.Response.Out.Header().Set("Status Code", "401 Authorization Required")
			// Only send the WWW-Authenticate header if this is not the TeamworkDesk app
			if twDeskVer := c.Request.Header.Get("twDeskVer"); twDeskVer == "" {
				c.Response.Out.Header().Set("WWW-Authenticate", `Basic realm="OhLife2"`)
			}
			return c.RenderText("401: Not authorized")
		}

		cache.Set(c.Session.Id()+"user", session, -1)

		c.Session.SetNoExpiration()

		// Just set it temporarily for the rest of this request
		c.User = session

		return nil
	}

	// CHECK USER LOGGING IN
	// No auth header sent up so just make sure the user is already logged in
	// and is not trying to log in.
	err := cache.Get(c.Session.Id()+"session", &session)
	if err != nil { //Not logged in...
		// HANDLE NO AUTH REQUESTS
		// Check if this is a request you don't have to be logged in for
		switch c.CanonRoute().Path {
		case "/v1/helpscout/created", "/v1/users/:userId/activate/:token", "/v1/tickets/:ticketId/solved/:token":

			c.User = &models.Session{
				User: models.User{},
			}

			return nil
		case "/v1/threads/:threadId/read.png", "/v1/customers/:customerId/fullcontact.json", "/v1/users/forgotpassword.json", "/v1/users/:userId/resetpassword.json", "/v1/login.json", "/v1/mailgun/notifymail", "/v1/mailgun/retryfailed", "/v1/mailgun/reprocessall", "/v1/settings/branding.json":
			return nil
		}

		// All other requests require user to be logged in - return 401 Authorization Required
		c.Response.Status = 401
		c.Response.Out.Header().Set("Status Code", "401 Authorization Required")
		// Only send the WWW-Authenticate header if this is not the TeamworkDesk app
		if twDeskVer := c.Request.Header.Get("twDeskVer"); twDeskVer == "" {
			c.Response.Out.Header().Set("WWW-Authenticate", `Basic realm="OhLife2"`)
		}
		return c.RenderText("401: Not authorized")
	}

	// All is good, the user is valid
	c.User = &session
	return nil
}

// Initialize a blank request object for the controller pointer at the
// start of every request
func (l *Authentication) LogInit() revel.Result {
	l.RequestLog = &modules.RequestLog{Queries: []modules.QueryLog{}}
	l.startTimeUnixNano = time.Now().UnixNano()
	modules.Requests[l.Controller] = l.RequestLog
	return nil
}

func (l *Authentication) LogExit() revel.Result {
	l.RequestLog.TimeTaken = (time.Now().UnixNano() - l.startTimeUnixNano) / 1000 / 1000

	if l.Controller.Params.Get("_profiling") != "" {
		for idx := range l.RequestLog.Queries {
			query := &l.RequestLog.Queries[idx]

			if params, ok := query.Params.([]interface{}); ok {
				if len(params) == 1 {
					if paramMap, k := params[0].(map[string]interface{}); k {
						for key, value := range paramMap {
							if str, conv := value.(string); conv {
								query.SQL = strings.Replace(query.SQL, ":"+key, "'"+str+"'", -1)
							} else if val, conv := value.(int64); conv {
								query.SQL = strings.Replace(query.SQL, ":"+key, strconv.FormatInt(val, 10), -1)
							}
						}
					}
				} else if len(params) > 1 {
					for _, value := range params {
						if str, conv := value.(string); conv {
							query.SQL = strings.Replace(query.SQL, "?", "'"+str+"'", 1)
						} else if val, conv := value.(int64); conv {
							query.SQL = strings.Replace(query.SQL, "?", strconv.FormatInt(val, 10), 1)
						}
					}
				}
			}
		}

		l.RenderArgs = map[string]interface{}{
			"queries":          l.RequestLog.Queries,
			"queryTimeTaken":   l.RequestLog.QueryTime,
			"queriesTakeUp":    (float64(l.RequestLog.QueryTime) / float64(l.RequestLog.TimeTaken)) * 100,
			"requestTimeTaken": l.RequestLog.TimeTaken,
		}

		return l.RenderTemplate("profiling.html")
	}

	l.RequestLog.User = fmt.Sprintf("%v", l.User)
	l.RequestLog.URL = l.Controller.Request.URL.String()
	if pathLogs, ok := modules.RequestLogs[l.CanonRoute().Path]; ok {
		pathLogs.MostRecent = (pathLogs.MostRecent + 1) % cap(pathLogs.RequestLogs)
		if len(pathLogs.RequestLogs) < cap(pathLogs.RequestLogs) {
			pathLogs.RequestLogs = append(pathLogs.RequestLogs, l.RequestLog)
		} else {
			pathLogs.RequestLogs[pathLogs.MostRecent] = l.RequestLog
		}
		pathLogs.NumCalls++
	}

	l.RenderArgs = map[string]interface{}{
		"queries":          l.RequestLog.Queries,
		"requestTimeTaken": l.RequestLog.TimeTaken,
		"queryTimeTaken":   l.RequestLog.QueryTime,
		"queriesTakeUp":    strconv.FormatFloat((float64(l.RequestLog.QueryTime)/float64(l.RequestLog.TimeTaken))*100, 'f', 2, 64),
	}

	delete(modules.Requests, l.Controller)
	return nil
}

// Returns the canonical route (e.g. "/chat/rooms/:chatRoomId/messages.json")
// of this request.
func (l *Authentication) CanonRoute() *revel.Route {
	method := l.Request.Method
	if method == "*" {
		method = ":METHOD"
	}

	leaf, _ := revel.MainRouter.Tree.Find("/" + method + l.Request.URL.Path)

	if route, ok := leaf.Value.(*revel.Route); ok {
		return route
	}
	return nil
}
