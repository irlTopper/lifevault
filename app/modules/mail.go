package modules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/irlTopper/lifevault/app/utility"
)

type EmailFields struct {
	Subject            string
	From               string
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
	Attachments        []string
}

type SMTPSettings struct {
	Server      string
	Username    string
	Password    string
	Port        int
	Security    string
	EmailFields *EmailFields
}

// Initialize the SMTP client
func InitSMTP() {
	fmt.Println("Config here")
}

var SendEmailTemplate = func(fields *EmailFields) error {

	fmt.Println("SendEmailTemplate", fields)

	return nil
}

func GetPlainTextVersionFromHTML(HTMLBody string, addReplyAboveLine bool) string {
	// Convert any newline type
	plaintextBody := regexp.MustCompile("(\\<\\/?((br)|(p))\\>)+").ReplaceAllString(HTMLBody, "\r\n\r\n")
	// Remove all tags but leave nice
	plaintextBody = utility.StripTags(plaintextBody)
	// Trim any extra space
	plaintextBody = strings.TrimSpace(plaintextBody)

	// Add on the please reply line to the plain text
	if addReplyAboveLine {
		plaintextBody = "-- Please reply above this line --\r\n" + plaintextBody
	}

	return plaintextBody
}
