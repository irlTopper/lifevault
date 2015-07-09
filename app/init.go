package app

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/irlTopper/lifevault/app/cron"
	"github.com/irlTopper/lifevault/app/filters"
	"github.com/irlTopper/lifevault/app/interceptors"
	"github.com/irlTopper/lifevault/app/models"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
	"github.com/revel/revel"
)

var ServerURL string // Global access to URL

func init() {

	//utilize all CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	revel.TypeBinders[reflect.TypeOf(map[int64]int64{})] = IntMapBinder

	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		filters.PanicFilter, // Recover from panics and display an error page instead.
		revel.RouterFilter,  // Use the routing table to select the right Action
		//revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,  // Parse parameters into Controller.Params.
		revel.SessionFilter, // Restore and write the session cookie.
		//revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter, // Restore kept validation errors and save new ones from cookie.
		//revel.I18nFilter,        // Resolve the requested language
		HeaderFilter,            // Add some security based headers
		revel.InterceptorFilter, // Run interceptors around the action.
		revel.CompressFilter,    // Compress the result.
		revel.ActionInvoker,     // Invoke the action.
	}

	revel.InterceptMethod((*interceptors.Authentication).LogInit, revel.BEFORE)
	revel.InterceptMethod((*interceptors.Authentication).LogExit, revel.AFTER)
	revel.InterceptMethod((*interceptors.Authentication).DecodeUserSession, revel.BEFORE)

	// Replacing the logger needs to be the first thing done
	revel.OnAppStart(func() {
		if revel.DevMode {
			logger.Log = logger.NewRevelLogger()
		} else {
			logger.Log = logger.NewLogglyLogger()
		}
	})
	revel.OnAppStart(ReportVersion)
	revel.OnAppStart(modules.InitDB)
	revel.OnAppStart(modules.InitSMTP)
	revel.OnAppStart(SetupDBTables)
	revel.OnAppStart(SetupStats)
	revel.OnAppStart(cron.InitCronJobs)

	modules.Requests = make(map[*revel.Controller]*modules.RequestLog)

}

func ReportVersion() {
	ServerURL, _ = revel.Config.String("server.URL")

	version := 1
	devMode := "PRODUCTION"
	if revel.DevMode {
		devMode = "DEVELOPER"
	}
	revel.INFO.Println("TeamworkDesk V%v running at %v in %v mode", version, ServerURL, devMode)
}

func SetupDBTables() {
	modules.DB.DbMap.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
	modules.DB.DbMap.AddTableWithName(models.MandrillMsg{}, "mandrillnotifications").SetKeys(true, "Id")
	modules.DB.DbMap.AddTableWithName(models.JournalEntry{}, "journalentries").SetKeys(true, "Id")
}

func SetupStats() {
	revel.MainRouter.Refresh()

	modules.RequestLogs = make(map[string]*modules.PathLog)
	for _, route := range revel.MainRouter.Routes {
		if strings.Contains(route.Path, "v1") {
			modules.RequestLogs[route.Path] = modules.NewPathLog(revel.Config.IntDefault("log.requests.num", 100))
		}
	}
}

var IntMapBinder = revel.Binder{
	Bind: func(params *revel.Params, name string, typ reflect.Type) reflect.Value {
		pValue := reflect.New(typ)
		return pValue.Elem()
	},
	Unbind: func(output map[string]string, name string, val interface{}) {
		revel.INFO.Println("Map:", output)
	},
}

// TODO turn this into revel.HeaderFilter
// should probably also have a filter for CSRF
// not sure if it can go in the same filter or not
var HeaderFilter = func(rc *revel.Controller, fc []revel.Filter) {
	// Add some common security headers
	rc.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	rc.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	rc.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](rc, fc[1:]) // Execute the next filter stage.
}
