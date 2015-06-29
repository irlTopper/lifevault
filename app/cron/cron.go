package cron

import (
	"fmt"

	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func InitCronJobs() {
	revel.INFO.Println("[CRON]: Initializing CRON jobs...")

	fmt.Println("[CRON]: Initializing CRON jobs...2")

	jobs.Schedule("@every 5m", SendDailyEmail{})
}

/*
	Tickets where we sent a message to the user but didn't hear a reply back
	withing 72 hours are automatically closed.

	This is a hack until we have proper triggers working.
*/
type SendDailyEmail struct {
	// filtered
}

func (e SendDailyEmail) Run() {
	revel.INFO.Println("Running 'SendDailyEmail' CRON job")

}
