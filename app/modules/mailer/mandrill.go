package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"path/filepath"
	"strings"

	"github.com/keighl/mandrill"
	"github.com/revel/revel"
)

var mandrillClient *mandrill.Client

func InitMandrill() {
	mandrillClient = mandrill.ClientWithKey(revel.Config.StringDefault("mandrill.key", ""))
	if mandrillClient == nil {
		panic(fmt.Sprintf("mandrillClient == nil"))
	}
}

type MandrillMailer struct {
}

func (self *MandrillMailer) Init(conf MailerConfig) {
	// noop
}

func (self *MandrillMailer) GetRenderedEmail() (html, text string) {
	return "", ""
}

func (self *MandrillMailer) Send(fields EmailFields) error {
	if mandrillClient == nil {
		InitMandrill()
	}
	message := &mandrill.Message{}

	// First we need to see if we need to render a template
	if fields.TemplateFile != "" && fields.HTMLBody == "" {

		if val, ok := fields.Data["RenderedHTML"]; ok && val != "" {
			fields.HTMLBody = val.(string)
		} else {
			var buff bytes.Buffer

			templateName := strings.Split(fields.TemplateFile, "/")

			t := template.New(templateName[len(templateName)-1])
			t.Funcs(revel.TemplateFuncs)
			t, err := t.ParseFiles(
				revel.ViewsPath+"/"+fields.TemplateFile,
				revel.ViewsPath+"/emails/powered.html",
			)

			if err != nil {
				return err
			}

			t.Funcs(revel.TemplateFuncs)

			err = t.Execute(&buff, fields.Data)

			if err != nil {
				return err
			}

			fields.HTMLBody = buff.String()
		}
	}

	// Set the body
	message.AutoText = true
	message.HTML = fields.HTMLBody
	message.Subject = fields.Subject
	message.FromName = fields.FromEmailName
	message.FromEmail = fields.From

	for name, email := range fields.To {
		if strings.TrimSpace(email) == "" {
			continue
		}

		if strings.TrimSpace(name) == "" {
			name = email
		}

		message.AddRecipient(email, name, "to")
	}

	for name, email := range fields.CC {
		message.AddRecipient(email, name, "cc")
	}

	for name, email := range fields.BCC {
		message.AddRecipient(email, name, "bcc")
	}

	// Add on any attachments
	for path, filename := range generateTempAttachments(fields.Attachments) {
		attachmentByte, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		base64Encoded := base64.StdEncoding.EncodeToString(attachmentByte)

		message.Attachments = append(message.Attachments, &mandrill.Attachment{
			Type:    mime.TypeByExtension(filepath.Ext(filename)),
			Name:    filename,
			Content: base64Encoded,
		})
	}

	// Initialize the headers map
	message.Headers = make(map[string]string)

	// If we don't have an active SPF record for this inbox, then
	// we should set the sender so it'll get a better delivery rating
	if fields.SenderRequired {
		message.Headers["Sender"] = strings.Replace(fields.From, "@", "=", 1) + "@lifevaultapp.com"
	}

	if fields.MessageId != "" {
		message.Headers["Message-Id"] = fields.MessageId
	}

	if fields.InReplyToMessageId != "" {
		message.Headers["In-Reply-To"] = fields.InReplyToMessageId
		message.Headers["References"] = fields.InReplyToMessageId
	}

	if fields.ReplyTo != "" {
		message.Headers["Reply-To"] = fields.ReplyTo
	}

	if fields.XBeenThere != "" {
		message.Headers["X-BeenThere"] = fields.XBeenThere
	}

	fmt.Println("SENDING EMAIL TO", fields.To, fields.Subject)

	// Send the email
	_, err := mandrillClient.MessagesSend(message)
	//logger.Log.Panicf("ERRORS: %s", err) // NOT SET
	if err != nil {
		fmt.Println("ERROR SENDING EMAIL", err)
		return err
	}
	return nil
}
