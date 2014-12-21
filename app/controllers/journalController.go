package controllers

import (
	"fmt"

	"github.com/irlTopper/ohlife2/app/interceptors"
	"github.com/revel/revel"
)

type CustomerController struct {
	*revel.Controller
	interceptors.Authentication
}

func (c CustomerController) PUT_v1_journalId(
	journalId int64,
	body string,
) revel.Result {

	fmt.Println("PUT_v1_journalId", journalId, body)

	return c.RenderJson(map[string]interface{}{"status": "ok"})
}

func (c CustomerController) GET_v1_journals_search(
	lastUpdated int64,
	search string,
	// Paging
	page int64,
	pageSize int64,
) revel.Result {

	fmt.Println("GET_v1_journals_search")

	return c.RenderJson(map[string]interface{}{"status": "ok"})
}

func (c CustomerController) GET_v1_journalId(journalId int64) revel.Result {

	fmt.Println("GET_v1_journalId")

	return c.RenderJson(map[string]interface{}{"status": "ok"})
}
