package cron

import (
	"fmt"
	"time"

	"github.com/irlTopper/lifevault/app/models"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/mailer"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func InitCronJobs() {
	revel.INFO.Println("[CRON]: Initializing CRON jobs...")

	fmt.Println("[CRON]: Initializing CRON jobs...2")

	jobs.Schedule("@every 1h", SendDailyEmail{})

	jobs.Now(SendDailyEmail{})
}

/*
	Tickets where we sent a message to the user but didn't hear a reply back
	withing 72 hours are automatically closed.

	This is a hack until we have proper triggers working.
*/

type Timezone struct {
	TimezoneId            int64
	TimezoneName          string
	TimezoneOffsetDisplay string
	TimezoneOffsetMins    int64
	TimezoneReferenceCode string
	currentHourOffset     int // Calculated
}

type User struct {
	Id         int64
	Email      string
	EmailToken string
	FirstName  string
	LastName   string
}

type SendDailyEmail struct{}

func (sde SendDailyEmail) Run() {
	//revel.INFO.Println("Running 'SendDailyEmail' CRON job")

	var err error

	// Optimize
	SQL := `SELECT 	timezones.timezoneId,
					timezones.timezoneName,
					timezones.timezoneOffsetDisplay,
					timezones.timezoneOffsetMins,
					timezones.timezoneReferenceCode
					FROM
					timezones
	`
	var timezones []Timezone
	_, err = modules.DB.Select(&revel.Controller{}, &timezones, SQL)
	if err != nil {
		fmt.Println("Failed to get time zone: ", err.Error())
		return
	}

	sentEmailCount := 0
	t := time.Now()
	for _, timezone := range timezones {
		// Get the current time in each time zone
		utc, err := time.LoadLocation(timezone.TimezoneReferenceCode)
		if err != nil {
			panic(fmt.Sprintf("Error looking timezone reference: %v, %v", timezone.TimezoneReferenceCode, err))
		}

		timeAtZone := t.In(utc)
		timezone.currentHourOffset = timeAtZone.Hour()

		// Get the users who are using this timezone and have selected to
		// get emails at this time of day
		var users []User
		SQL := `SELECT 	users.id,
						users.email,
						users.emailToken,
						users.firstName,
						users.lastName
				FROM 	users
				WHERE 	timezoneId = :timezoneId
						AND dailyEmailReminderTime >= :hour
						AND dailyEmailReminderTime < (:hour + 1)
						AND status = 'active'
			`
		_, err = modules.DB.Select(&revel.Controller{}, &users, SQL, map[string]interface{}{
			"timezoneId": timezone.TimezoneId,
			"hour":       timezone.currentHourOffset,
		})
		if err != nil {
			fmt.Println("Failed to get users: ", err.Error())
			return
		}

		if len(users) == 0 {
			continue
		}

		todayInCurrentTimezone := timeAtZone.Format("Monday, Jan _2")
		subject := "It's " + todayInCurrentTimezone + " - How did your day go?"

		for _, user := range users {
			sde.SendMailForUser(&user, subject)
			sentEmailCount = sentEmailCount + 1
		}
	}

	if sentEmailCount == 0 {
		fmt.Println("No matching users for this hour")
	}

}

func (sde SendDailyEmail) SendMailForUser(user *User, subject string) {

	var err error

	from := user.EmailToken + "@lifevaultapp.com"

	fmt.Println("---------------------------------------------")
	fmt.Println("Sending email")
	fmt.Println("To:", user.Email)
	fmt.Println("From ", from)
	fmt.Println(subject)

	HTMLBody := "Just reply to this email with your entry."
	textBody := "Just reply to this email with your entry."

	// Pick a previous random entry
	SQL := `SELECT 	date, body
			FROM 	journalentries
			WHERE 	users_id = ?
					AND state = 'published'
			ORDER BY rand()
			limit 1`
	journalEntry := models.JournalEntry{}
	err = modules.DB.SelectOne(&revel.Controller{}, &journalEntry, SQL, user.Id)
	if err == nil {
		HTMLBody += "<br/><br/>Remember this? A while back you wrote:<br/><br/>" + journalEntry.Body
	}

	err = modules.SendEmail(modules.SendEmailData{
		Email: &mailer.EmailWithConfig{Fields: mailer.EmailFields{
			Subject:       subject,
			From:          from,
			HTMLBody:      HTMLBody,
			PlaintextBody: textBody,
			To:            map[string]string{user.FirstName + " " + user.LastName: user.Email},
		}},
	})
	if err != nil {
		fmt.Println("Error sending email", err)
	}
}
