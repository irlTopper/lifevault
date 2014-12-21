package controllers

import (
	"time"

	"encoding/json"

	"github.com/irlTopper/ohlife2/app/interceptors"
	"github.com/irlTopper/ohlife2/app/modules"
	"github.com/revel/revel"
)

type AppController struct {
	*revel.Controller
	interceptors.Authentication
}

func (c AppController) Index() revel.Result {
	return c.RenderText("Teamwork API")
}

func (c AppController) HandleCORS() revel.Result {
	if origin := c.Request.Header.Get("Origin"); origin != "" {
		c.Response.Out.Header().Set("Access-Control-Allow-Origin", origin)
	}

	c.Response.Out.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Response.Out.Header().Set("Access-Control-Allow-Headers", "Set-Cookie, Cookie, Authorization, X-Page, X-Pages, X-Records, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	c.Response.Out.Header().Set("Access-Control-Expose-Headers", "Set-Cookie, Cookie, Authorization, X-Page, X-Pages, X-Records, X-lastUpdated, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	c.Response.Out.Header().Set("Access-Control-Allow-Credentials", "true")

	c.Response.Out.WriteHeader(204)

	return nil
}

func (c AppController) GET_version() revel.Result {
	return c.RenderJson(map[string]interface{}{"version": 0.2})
}

type PanicLog struct {
	Id        int64
	Data      string
	CreatedAt time.Time
}

func (c AppController) POST_panics() revel.Result {
	return c.Redirect("/panics")
}

func (c AppController) GET_panics() revel.Result {
	SQL := `
	SELECT
		id,
		data,
		createdAt
	FROM paniclogs
	ORDER BY createdAt DESC`

	var panics []PanicLog

	modules.DB.Select(c.Controller, &panics, SQL)

	panicData := []map[string]interface{}{}

	for _, p := range panics {
		var data map[string]interface{}

		json.Unmarshal([]byte(p.Data), &data)

		data["time"] = p.CreatedAt
		data["id"] = p.Id

		if d, ok := data["title"].(map[string]interface{}); ok {
			data["title"] = d["Message"]
		}

		panicData = append(panicData, data)
	}

	c.RenderArgs = map[string]interface{}{
		"panics": panicData,
	}

	return c.RenderTemplate("errors/paniclist.html")
}

func (c AppController) GET_panicId(panicId int64) revel.Result {
	SQL := `
	SELECT
		id,
		data,
		createdAt
	FROM paniclogs
	WHERE id = ?`

	var p PanicLog

	err := modules.DB.SelectOne(c.Controller, &p, SQL, panicId)

	if err != nil {
		return c.RenderText("Panic with that ID not found..")
	}

	var data map[string]interface{}

	json.Unmarshal([]byte(p.Data), &data)

	c.RenderArgs = data

	return c.RenderTemplate("errors/panic.html")
}
