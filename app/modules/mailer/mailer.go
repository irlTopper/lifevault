package mailer

import (
	"bytes"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/irlTopper/lifevault/app/modules/aws"
	"github.com/irlTopper/lifevault/app/utility"
	"github.com/revel/revel"
)

type EmailFields struct {
	Subject            string
	From               string
	FromEmailName      string
	Name               string
	PlaintextBody      string
	HTMLBody           string
	To                 map[string]string
	MessageId          string
	InReplyToMessageId string
	ReplyTo            string
	CC                 map[string]string
	BCC                map[string]string
	TemplateFile       string
	Data               map[string]interface{}
	Attachments        []Attachment
	TicketId           sql.NullInt64
	IsResend           bool
	XBeenThere         string
	SenderRequired     bool
	Type               string
}

type Attachment struct {
	Id           int64  `json:"id"`
	FileName     string `json:"filename"`
	DownloadURL  string `json:"downloadurl"`
	MimeType     string `json:"mimetype"`
	Size         int64  `json:"size"`
	ThumbnailURL string `json:"thumbnailURL"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
	ThreadId     int64  `json:"-"`
	S3Path       string `json:"-"`
}

type MailerConfig struct {
	Server   string
	Username string
	Password string
	Port     int
	Security string
	Provider string
}

type EmailWithConfig struct {
	Config MailerConfig
	Fields EmailFields
}

type Mailer interface {
	Init(MailerConfig)
	Send(EmailFields) error
	GetRenderedEmail() (html, text string)
}

func generateTempAttachments(attachments []Attachment) (returnAttachments map[string]string) {
	// Initialize the map first
	returnAttachments = map[string]string{}

	for _, attachment := range attachments {
		path := utility.TempFolder + strconv.FormatInt(attachment.Id, 10) + "_" + attachment.FileName

		downloadURL := attachment.DownloadURL
		if attachment.S3Path != "" {
			downloadURL = aws.AWS.S3.GetFileURL(attachment.S3Path, attachment.FileName, false)
		}

		// First we check if we don't already have this file in temp
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_, err = utility.DownloadFromURL(downloadURL, utility.TempFolder, strconv.FormatInt(attachment.Id, 10)+"_"+attachment.FileName)

			if err != nil {
				panic(err)
			}
		}

		returnAttachments[path] = attachment.FileName
	}

	return
}

func BuildTemplate(templateFile string, data map[string]interface{}) (string, error) {
	var buff bytes.Buffer
	templateName := strings.Split(templateFile, "/")

	t := template.New(templateName[len(templateName)-1])
	t.Funcs(revel.TemplateFuncs)
	t, err := t.ParseFiles(
		revel.ViewsPath+"/"+templateFile,
		revel.ViewsPath+"/emails/powered.html",
	)

	if err != nil {
		return "", err
	}

	t.Funcs(revel.TemplateFuncs)

	err = t.Execute(&buff, data)

	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func getPlainTextVersionFromHTML(HTMLBody string, addReplyAboveLine bool) string {
	plaintextBody := utility.StripTagsWithNewLines(HTMLBody)

	// Add on the please reply line to the plain text
	if addReplyAboveLine {
		plaintextBody = "-- Please reply above this line --\r\n" + plaintextBody
	}

	return plaintextBody
}
