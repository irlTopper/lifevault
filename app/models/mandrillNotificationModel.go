package models

import (
	"errors"
	"strings"
	"time"

	"github.com/revel/revel"

	"github.com/go-gorp/gorp"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
)

var (
	ErrArchiveDecompression = errors.New("Error decompressing email archive")
)

const (
	ValidEmailReceiveDomain            = "lifevaultapp.com"
	SpamThreshold              float64 = 5.0
	ObviousSpamThreshold       float64 = 8
	ArchiveDecompressStartDate int64   = 1436227200 // The date that we switched to the new archive database.
)

const (
	IsNotSpam_Id      = 0
	IsObviousSpam_Id  = 1
	IsProbablySpam_Id = 2
)

type MandrillStats struct {
	PostCounter int64 `json:"counter"`
}

type MandrillEvent struct {
	Ts    int64
	Event string
	Msg   MandrillMsg `json:"msg"`
}

type MandrillMsg struct {
	Id         int64
	Raw_msg    string                 `db:"raw_msg" json:"raw_msg"`
	Headers    map[string]interface{} `db:"-" json:"headers"`
	Text       string                 `db:"-"`
	Html       string                 `db:"HTML`
	From_email string                 `db:"from_email"`
	From_name  string                 `db:"from_name"`
	To         []interface{}          `db:"-"`
	Email      string                 `db:"email"`
	Subject    string                 `db:"subject"`
	State      string                 `db:"state"`
	Tags       interface{}            `db:"-"`
	Sender     string                 `db:"sender"`
	//Attachments []MandrillEmailAttachment `db:"-"`
	//Images      []MandrillEmailAttachment `db:"-"`
	Spam_report MandrillSpamReport `db:"-"`
	// Extra
	RetryCount      int64  `db:"retryCount"`
	ProcessingNotes string `db:"processingNotes"`
}

type MandrillEmailAttachment struct {
	Name    string
	Type    string
	Content string
	Base64  bool
}

type MandrillSpamReport struct {
	Score         float32
	matched_rules []interface{}
}

type MandrillMsgArchive struct {
	MandrillMsg
	Body MandrillMsgBody `db:"-"` // The actual compressed data
}

type MandrillMsgBody struct {
	Id           int64     `db:"id"`
	ArchiveId    int64     `db:"archive_id"`
	BodyHTML     []byte    `db:"bodyHTML"`
	BodyPlain    []byte    `db:"bodyPlain"`
	StrippedHTML []byte    `db:"strippedHTML"`
	StrippedText []byte    `db:"strippedText"`
	CreatedAt    time.Time `db:"createdAt"`
}

func NotificationById(id int64, session *Session) (notification *MandrillMsgArchive, err error) {
	SQL := `
		SELECT
			mandrillnotifications.id,
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
		WHERE id = ?
	`

	err = modules.DB.SelectOne(&revel.Controller{}, &notification, SQL, id, id)
	if err != nil {
		logger.Log.Printf("[MANDRILL] Unable to find %d in old archive.", id)
	}

	return notification, err
}

type MandrillMsgFile struct {
	Id                int64
	MandrillMsgId     int64 `db:"mandrillnotifications_id"`
	LocalFilePath     string
	FileName          string
	OriginalFileName  string
	OriginalFieldName string
	Size              int64
}

func (email *MandrillMsg) PreInsert(s gorp.SqlExecutor) error {
	type SimpleEmail struct {
		From    string
		Subject string
	}
	logger.Log.Printf("[MANDRILL] Email received: %+v", SimpleEmail{
		From:    email.Sender,
		Subject: email.Subject,
	})
	return nil
}

func (email *MandrillMsg) PostUpdate(s gorp.SqlExecutor) error {
	return nil
}

func (email *MandrillMsg) ProcessAttachments(controller *revel.Controller, session *Session) error {
	/*
		for i := 1; i <= email.AttachmentCount; i++ {
			fieldName := "attachment-" + strconv.Itoa(i)

			// Process the file into the data and save it to disk:
			tempFileRefs, err := SaveUploadedFilesToTemp(TempFileOptions{
				FieldName: fieldName,
			},
				controller,
				session,
			)
			if err != nil {
				return err
			}
			if len(tempFileRefs) != 1 {
				return errors.New("Only one file should be uploaded")
			}

			tempFileRef := tempFileRefs[0]

			// Save this mailfunnotificationfile to the database
			mandrillNotificationFile := MandrillMsgFile{
				MandrillMsgId: email.Id,
				LocalFilePath:          tempFileRef.LocalFilePath,
				FileName:               tempFileRef.FileName,
				OriginalFileName:       tempFileRef.OriginalFileName,
				OriginalFieldName:      tempFileRef.OriginalFieldName,
				Size:                   tempFileRef.Size,
			}
			modules.DB.Insert(controller, &mandrillNotificationFile)

			// Append this file to the temp files refs in the "email" object so we can process it now
			email.UploadFilesRefs = append(email.UploadFilesRefs, tempFileRef)
		}
	*/
	return nil
}

// Validate the the recipient email looks like emailToken@lifevaultapp.com
func (mail *MandrillMsg) IsValidReplyAddress() error {
	parts := strings.Split(mail.Email, "@")

	if len(parts) != 2 {
		return errors.New("Recipient '" + mail.Email + "' doesn't look like a valid forwarding email address for TeamworkDesk (1)")
	}
	if parts[1] != ValidEmailReceiveDomain {
		return errors.New("Recipient '" + mail.Email + "' doesn't look like a valid forwarding email address for TeamworkDesk (2)")
	}

	return nil
}
