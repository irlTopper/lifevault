package cron

import (
	"fmt"
	"time"

	"github.com/irlTopper/lifevault/app/modules"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func InitCronJobs() {
	revel.INFO.Println("[CRON]: Initializing CRON jobs...")

	fmt.Println("[CRON]: Initializing CRON jobs...2")

	jobs.Schedule("@every 2s", SendDailyEmail{})
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
	Id        int64
	Email     string
	FirstName string
	LastName  string
}

type SendDailyEmail struct{}

func (sde SendDailyEmail) Run() {
	revel.INFO.Println("Running 'SendDailyEmail' CRON job")

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

	t := time.Now()
	for _, timezone := range timezones {
		// Get the current time in each time zone
		utc, err := time.LoadLocation(timezone.TimezoneReferenceCode)
		if err != nil {
			panic(fmt.Sprintf("Error looking timezone reference: %v, %v", timezone.TimezoneReferenceCode, err))
		}

		timezone.currentHourOffset = t.In(utc).Hour()

		// Get the users who are using this timezone and have selected to
		// get emails at this time of day
		var users []User
		SQL := `SELECT 	users.id,
							users.email,
							users.firstName,
							users.lastName
							FROM
							users
							WHERE dailyEmailReminderTime >= :hour AND dailyEmailReminderTime < (:hour + 1)
							AND status = 'active'
			`
		_, err = modules.DB.Select(&revel.Controller{}, &users, SQL, map[string]interface{}{
			"hour": timezone.currentHourOffset,
		})
		if err != nil {
			fmt.Println("Failed to get users: ", err.Error())
			return
		}

		for _, user := range users {
			sde.SendMailForUser(&user)
		}
	}

}

func (sde SendDailyEmail) SendMailForUser(user *User) {
	fmt.Println("Sending email to ", user.Email)

}
