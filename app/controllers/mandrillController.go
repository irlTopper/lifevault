// NoDoc
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/irlTopper/lifevault/app/filters"
	"github.com/irlTopper/lifevault/app/interceptors"
	"github.com/irlTopper/lifevault/app/models"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
	"github.com/revel/revel"
	revelCache "github.com/revel/revel/cache"
)

type MandrillController struct {
	*revel.Controller
	interceptors.Authentication
}

type MsgRetry struct {
	Msg        *models.MandrillMsg
	Controller *MandrillController
}

func (retry MsgRetry) Run() {
	msg := models.MandrillProcessor{Msg: retry.Msg}
	err := msg.Process(retry.Controller.Controller)
	if err != nil {
		retry.Controller.EmailProcessingFailed(retry.Msg, err.Error())
	}
}

// NoDoc
func (c MandrillController) GET_v1_mandrill_process_pending() revel.Result {
	// Get the next failed one out of the database
	var emails []models.MandrillMsg
	SQL := `
		SELECT mandrillnotifications.id,
			mandrillnotifications.attachmentCount,
			mandrillnotifications.bodyPlain,
			mandrillnotifications.bodyHTML,
			mandrillnotifications.contentIdMap,
			mandrillnotifications.from,
			mandrillnotifications.messageHeaders,
			mandrillnotifications.messageURL,
			mandrillnotifications.recipient,
			mandrillnotifications.sender,
			mandrillnotifications.subject,
			mandrillnotifications.strippedText,
			mandrillnotifications.strippedSignature,
			mandrillnotifications.strippedHTML,
			mandrillnotifications.signature,
			mandrillnotifications.token,
			mandrillnotifications.timestamp,
			mandrillnotifications.processingNotes,
			mandrillnotifications.state,
			mandrillnotifications.retryCount
		FROM mandrillnotifications
		WHERE state = 'pending'
		ORDER BY id
	`
	modules.DB.Select(c.Controller, &emails, SQL)

	for _, email := range emails {
		go func(msg models.MandrillMsg) {
			defer filters.CatchThreadedPanics()
			p := models.MandrillProcessor{Msg: &msg}
			p.Process(c.Controller)
		}(email)
	}

	return c.RenderText("Queued " + strconv.Itoa(len(emails)) + " pending emails for processing!")
}

// GET /v1/mandrill/notifymail
// Just returns some stats
func (c MandrillController) GET_v1_mandrill_notifymail() revel.Result {
	var stats models.MandrillStats
	revelCache.Get("mandrillMailsReceived", &stats)
	return c.RenderJson(map[string]interface{}{"status": "ok", "stats1": stats.PostCounter})
}

func (c MandrillController) POST_v1_mandrill_notifymail() revel.Result {

	var err error

	// Update the stats
	var stats models.MandrillStats
	revelCache.Get("mandrillMailsReceived", &stats)
	stats.PostCounter++
	revelCache.Set("mandrillMailsReceived", &stats, -1)

	var mandrill_eventsString string = c.Request.Form.Get("mandrill_events")

	//var mandrill_events interface{}
	var mandrill_events []*models.MandrillEvent
	err = json.Unmarshal([]byte(mandrill_eventsString), &mandrill_events)
	if err != nil {
		return modules.JSONError(c.Controller, http.StatusNotAcceptable, err.Error())
	}

	for i := range mandrill_events {
		mandrillEvent := mandrill_events[i]

		mandrillEvent.Msg.State = "pending"

		// Save this to the database
		err = modules.DB.Insert(c.Controller, &mandrillEvent.Msg)
		if err != nil {
			logger.Log.Panicf("[MANDRILL] Message insertion failed: %s", err.Error())
		}

		if err != nil {
			return modules.JSONError(c.Controller, http.StatusInternalServerError, err.Error())
		}

		// Once the insert has happened and we have added any attachments
		// let's move to a background thread for processing the e-mail

		c.processNotification(&mandrillEvent.Msg)

		//TODO
		/*
			go func(email models.MandrillMsg) {
				defer filters.CatchThreadedPanics()
				c.processNotification(&MandrillEvent.Msg)
			}(MandrillEvent.Msg)
		*/
	}

	// If we were able to insert into the database, return a success.
	// mandrill doesn't need to care about our own internal processing.
	return c.RenderJson(map[string]interface{}{
		"status": "ok",
	})
}

func (c MandrillController) processNotification(msg *models.MandrillMsg) revel.Result {
	p := models.MandrillProcessor{Msg: msg}

	fmt.Println("Processed message", msg.Id)

	// Mark the email as having started processing
	msg.State = "started"
	modules.DB.Update(c.Controller, msg)

	var err error
	err = p.Process(c.Controller)

	if err != nil {
		// logger.Log.Panicf("[MAIL] Error processing email ID %d; Error was %s", email.Id, err.Error())
		return c.EmailProcessingFailed(msg, err.Error())
	} else {
		// Set the state to processed-ok
		modules.DB.Exec(c.Controller, "UPDATE mandrillnotifications SET state = 'processed-ok' WHERE id = ? LIMIT 1", msg.Id)
		return c.EmailProcessingSuccess()
	}
}

type HTML string

func (r HTML) Apply(req *revel.Request, resp *revel.Response) {
	resp.WriteHeader(http.StatusOK, "text/html")
	resp.Out.Write([]byte(r))
}

func (c MandrillController) EmailProcessingSuccess() revel.Result {
	return c.RenderJson(map[string]interface{}{"status": "ok"})
}

func (c MandrillController) EmailProcessingFailed(notification *models.MandrillMsg, processingNotes string) revel.Result {
	// If a notification process failed, we need to re-run this ASAP.
	if notification.RetryCount < models.MaxRetryCount {
		// retryIn := time.Duration(notification.RetryCount+1) * time.Minute
		// jobs.In(retryIn, NotificationRetry{Notification: notification, Controller: &c})
	}

	return modules.JSONError(c.Controller, http.StatusInternalServerError, processingNotes)
}
