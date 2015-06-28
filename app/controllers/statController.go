package controllers

import (
	"fmt"

	"github.com/revel/revel"

	"github.com/teamwork/TeamworkDesk/app/interceptors"
	"github.com/teamwork/TeamworkDesk/app/modules"

	"runtime"
)

type StatController struct {
	*revel.Controller
	interceptors.Authentication
}

// DeleteStats resets the request logs.
func (c *StatController) DeleteStats() revel.Result {
	for path := range modules.RequestLogs {
		modules.RequestLogs[path] =
			modules.NewPathLog(revel.Config.IntDefault("log.requests.num", 100))
	}
	return c.Stats("")
}

// subdomain is used to populate the subdomains of teamwork.com in the stats
// page.
type subdomain struct {
	Id   int // This is just for populating the id attributes in stats.html
	Name string
}

func (c *StatController) Stats(view string) revel.Result {
	if view != "" {
		return c.statsSingle(view)
	}

	paths := make([]string, 0, len(modules.RequestLogs))
	for path := range modules.RequestLogs {
		paths = append(paths, path)
	}

	subdomains := make([]subdomain, 0, 6)
	for i := 0; i < cap(subdomains); i++ {
		subdomains = append(subdomains, subdomain{
			Id:   i + 1,
			Name: fmt.Sprintf("appserver%d", i+1),
		})
	}

	c.RenderArgs = map[string]interface{}{
		"developer":  true,
		"paths":      paths,
		"subdomains": subdomains,
	}

	return c.RenderTemplate("stats.html")
}

type Request struct {
	Queries   []modules.QueryLog
	User      string // Session
	URL       string
	TimeTaken int64
	Cache     bool
	ETag      bool
}

// statsSingle creates a page for view statistics for a single URL.
func (c *StatController) statsSingle(view string) revel.Result {
	var logs []*modules.RequestLog
	if path, ok := modules.RequestLogs[view]; ok {
		if len(path.RequestLogs) < cap(path.RequestLogs) {
			logs = path.RequestLogs
		} else {
			// Slices are used here to put the logs into the correct order.
			logs = append(path.RequestLogs[path.MostRecent:], path.RequestLogs[:path.MostRecent]...)

		}
	} else {
		logs = make([]*modules.RequestLog, 0, 0)
	}

	requests := make([]*Request, 0, len(logs))
	for i := 0; i < len(logs); i++ {
		requests = append(requests, &Request{
			Queries:   logs[i].Queries,
			User:      logs[i].User,
			URL:       logs[i].URL,
			TimeTaken: logs[i].TimeTaken,
			Cache:     false,
			ETag:      false,
		})
	}

	c.RenderArgs = map[string]interface{}{
		"path":     view,
		"requests": requests,
		"logSize":  len(requests),
	}

	return c.RenderTemplate("stats_single.html")
}

func (c *StatController) StatsJson() revel.Result {
	return c.RenderJson(map[string]interface{}{
		"MemoryInUse": runtime.MemStats{}.Alloc / 1024 / 1024,
		"Version":     1,
		"aaData":      c.stats(),
	})
}

type StatsJSON struct {
	Name    string
	Called  int64
	Average int64
	Total   int64
	Cached  string
	ETagged string
}

func (c *StatController) stats() []*StatsJSON {
	var stats []*StatsJSON
	for path, logs := range modules.RequestLogs {
		var total int64
		for _, log := range logs.RequestLogs {
			total += log.TimeTaken
		}

		var average int64
		if len(logs.RequestLogs) != 0 {
			average = total / int64(len(logs.RequestLogs))
		}

		stats = append(stats, &StatsJSON{
			Name:    path,
			Called:  logs.NumCalls,
			Average: average,
			Total:   total,
			Cached:  "0 (0%)",
			ETagged: "0 (0%)",
		})
	}
	return stats
}
