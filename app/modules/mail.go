package modules

import (
	"fmt"

	"github.com/irlTopper/lifevault/app/modules/logger/app"
	"github.com/irlTopper/lifevault/app/modules/mailer"
)

var (
	mandrillMailer mailer.Mailer
)

// Initialize the SMTP client
func InitSMTP() {
	mailer.InitMandrill()
	mandrillMailer = &mailer.MandrillMailer{}
}

type EmailSendError struct {
	Options     *mailer.EmailFields
	Message     string
	MessageType string
}

type EmailSendErrFunc func(EmailSendError)

var NoopEmailErr = func(errInfo EmailSendError) {
	// Note that with the revel logger we are not actually
	// going to panic here, despite the name.
	logger.Log.Panicf("[MAIL] Error sending email %s", errInfo.Message)
	return
}

var EmailErrorLogger = func(errInfo EmailSendError) {

}

type SendEmailData struct {
	Email       *mailer.EmailWithConfig
	Sender      mailer.Mailer
	ErrFunc     EmailSendErrFunc
	MessageType string
}

func (self *SendEmailData) ApplyDefaults() {
	if self.ErrFunc == nil {
		self.ErrFunc = NoopEmailErr
	}

	if self.MessageType == "" {
		self.MessageType = "normalEmail"
	}

	// For all outbound e-mail on local machines we are going to send
	// email through mailtrap (by default) and only send to the dev team rather than real customers.
	/*
		if revel.DevMode {
			self.Sender = mailer.NewDevMailer()
		} else {
			self.Sender.Init(self.Email.Config)
		}
	*/
	self.Sender = mandrillMailer
	self.Sender.Init(self.Email.Config)
}

func (self *SendEmailData) logEmailAsSent() {
	/*
		ticketData, err := json.Marshal(map[string]interface{}{
			"to":        self.Email.Fields.To,
			"from":      self.Email.Fields.From,
			"cc":        self.Email.Fields.CC,
			"bcc":       self.Email.Fields.BCC,
			"subject":   self.Email.Fields.Subject,
			"messageId": self.Email.Fields.MessageId,
		})
		if err != nil {
			logger.Log.Panicf("[MAIL] Failed to marshal json for email data")
		} else {
			logger.Log.Printf("[MAIL] Delivered Success: %s", string(ticketData))
		}
	*/
	return
}

func SendEmail(message SendEmailData) error {
	message.ApplyDefaults()

	err := message.Sender.Send(message.Email.Fields)

	if err != nil {
		fmt.Println("X2. Error sending email", err, message.Email.Fields)
	}
	/*
		if err != nil {
			emailData, err2 := json.Marshal(message.Email.Fields)
			if err2 == nil {
				logger.Log.Panicf("[MAIL] Error sending email: %s : Email Data %s", err.Error(), string(emailData))
			}

			// get the rendered email, save re-doing the template during a resend
			html, text := message.Sender.GetRenderedEmail()

			if message.Email.Fields.Data != nil {
				message.Email.Fields.Data["RenderedHTML"] = html
				message.Email.Fields.Data["RenderedText"] = text
			}

			message.ErrFunc(EmailSendError{
				Options:     &message.Email.Fields,
				Message:     err.Error(),
				MessageType: message.MessageType,
			})
			return err
		}
	*/

	message.logEmailAsSent()

	// Remove any attachments
	/*
		for _, attachment := range message.Email.Fields.Attachments {
			os.Remove(utility.TempFolder + strconv.FormatInt(attachment.Id, 10) + "_" + attachment.FileName)
		}
	*/

	return nil
}
