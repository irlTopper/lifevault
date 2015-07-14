package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
	"github.com/revel/revel"
)

const (
	// The robot user is responsible for processing
	// new tickets and creating threads.  Used in
	// mandrill controller.
	MaxRetryCount = 5
)

var RobotUser = User{
	Id:        RobotUserId,
	Email:     "noreply@lifevaultapp.com",
	FirstName: "LifeVault",
}

var (
	InvalidTokenError = errors.New("Unable to process email as it used an invalid token")
)

type MandrillProcessor struct {
	Msg *MandrillMsg
}

type EmailMessageIdLookup struct {
	InstallationId int64
	TicketId       int64
	MessageId      string
	EmailToken     string
}

// Process the email!
// TODO: Extract each section of code to its own method
// for easier regression testing
func (p MandrillProcessor) Process(controller *revel.Controller) error {
	var (
		err                 error
		emailToken          string
		emailTokenAndDomain []string
	)

	// First step, check if this is an e-mail from ourselves.
	// If an agent has an e-mail that redirects to the inbox, sending out a notification
	// to that e-mail is going to result in an infinite loop, which we don't want.
	// So let's check if we sent the e-mail.  If we did, just mark it
	// as processed, add a note, and move on.
	if p.Msg.Sender == "notifications@lifevaultapp.com" {
		p.EmailProcessingSuccess(controller, "Preventing redirect loop (sent from lifeVault)")
		return nil
	}

	err = p.Msg.IsValidReplyAddress()
	if err != nil {
		p.EmailProcessingFailed(controller, "Couldn't find out where this email is supposed to go!")
		return errors.New("Couldn't find out where this email belongs..")
	}

	emailTokenAndDomain = strings.Split(p.Msg.Email, "@")
	emailToken = emailTokenAndDomain[0]

	// Find the matching user
	var matchedUser User
	SQL := `SELECT id
			FROM users
			WHERE emailToken = ?
			limit 1`
	err = modules.DB.SelectOne(controller, &matchedUser, SQL, emailToken)
	if err != nil {
		p.EmailProcessingFailed(controller, "No matching user for this email token")
		return errors.New("No matching user for this email token: " + emailToken)
	}
	fmt.Println("Found user", matchedUser)

	body := p.GetCleanJournalBody()

	// Insert the journal into the database
	journalEntry := JournalEntry{
		Users_id: matchedUser.Id,
		Date:     time.Now(),
		Body:     body,
	}
	err = modules.DB.Insert(controller, &journalEntry)
	if err != nil {
		logger.Log.Panicf("journalEntry insertion failed: %s", err.Error())
	}

	//socketio.BroadcastMessage(installation.Id, "updateNotice", map[string]interface{}{"updateName": "updateInboxes", "inboxId": dbInbox.Id, "userId": session.Id})
	p.EmailProcessingSuccess(controller, "Email was successfully processed")
	//logger.Log.Printf("[MAIL] Process returning. L1000")

	return nil
}

func (p MandrillProcessor) EmailProcessingSuccess(controller *revel.Controller, status string) {
	logger.Log.Printf("[MAIL] Email proccessed ok")
	p.Msg.State = "processed-ok"
	p.Msg.ProcessingNotes = status
	modules.DB.Update(controller, p.Msg)
}

func (p MandrillProcessor) EmailProcessingFailed(controller *revel.Controller, status string) {
	logger.Log.Printf("[MAIL] Marking as failed")
	if p.Msg.RetryCount == 0 {
		p.Msg.State = "processed-failed"
	} else {
		p.Msg.State = "processed-failedRetry"
	}

	p.Msg.RetryCount += 1
	p.Msg.ProcessingNotes = status
	_, err := modules.DB.Update(controller, p.Msg)
	if err != nil {
		panic(err.Error())
	}
}

func (p MandrillProcessor) GetCleanJournalBody() (body string) {

	BodyPlain := p.Msg.Text
	BodyHTML := p.Msg.Html
	// We don't want to use the new processing on agent emails just yet for safety

	// Check if it's a plaintext email
	if strings.TrimSpace(BodyHTML) == "" {

		body = modules.StripPlainEmailReplies(BodyPlain)
		_, body = modules.SplitPlainEmailBodyAndSignature(BodyPlain, body)

		if strings.TrimSpace(body) == "" {
			body = BodyPlain
		}

		body = modules.FormatPlainEmail(body)

	} else {

		fmt.Println("here001", len(BodyHTML))

		_, body = modules.SplitHTMLEmailBodyAndSignature("", BodyHTML, true)

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

		if strings.TrimSpace(doc.Text()) == "" {
			fmt.Println("here002")
			body = BodyHTML
		}

		body = modules.CleanupHTMLEmail(body)

	}
	return
}
