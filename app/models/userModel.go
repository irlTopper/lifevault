package models

import (
	"time"

	"github.com/coopernurse/gorp"
	"github.com/guregu/null/zero"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/revel/revel"
)

const (
	// The robot user is responsible for processing
	// new tickets and creating threads.  Used in
	// mandrill controller.
	RobotUserId = 999999999
)

type User struct {
	// Identity fields
	Id int64 `json:"id"`
	// Standard fields A-Z
	AutoLoginCode          string `json:"-"`
	CreatedAt              time.Time
	Email                  string   `json:"email"`
	FirstName              string   `json:"firstName"`
	LastName               string   `json:"lastName"`
	Password               string   `json:"-" db:"password"`
	Salt                   zero.Int `json:"-" db:"passwordSalt"`
	TimeZoneId             int64    `json:"timezoneId" db:"timezoneId"`
	UpdatedAt              time.Time
	VisitCount             int64 `json:"-"`
	DailyEmailReminderTime int64 `json:"dailyEmailReminderTime"`
	// State fields
	ZuluCreatedAt string `json:"createdAt"`
	ZuluUpdatedAt string `json:"updatedAt"`
	State         string `json:"-"`
}

func (i *User) PreInsert(s gorp.SqlExecutor) error {
	i.CreatedAt = time.Now().UTC()
	i.UpdatedAt = i.CreatedAt
	return nil
}

func (i *User) PreUpdate(s gorp.SqlExecutor) error {
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func IsValidUser(userId int64, rc *revel.Controller, session *Session) bool {
	id, _ := modules.DB.SelectInt(rc, "SELECT userId FROM users WHERE userId = ?", userId)
	return id != 0
}
