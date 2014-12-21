package models

import (
	"time"

	"github.com/coopernurse/gorp"
	"github.com/guregu/null/zero"
	"github.com/irlTopper/ohlife2/app/modules"
	"github.com/revel/revel"
)

const (
	// The robot user is responsible for processing
	// new tickets and creating threads.  Used in
	// mailgun controller.
	RobotUser = 999999999
)

type User struct {
	// Identity fields
	Id int64 `json:"id"`
	// Standard fields A-Z
	FirstName     string   `json:"firstName"`
	LastName      string   `json:"lastName"`
	AutoLoginCode string   `json:"-"`
	Email         string   `json:"email"`
	TimeZoneId    int64    `json:"timezoneId"`
	VisitCount    int64    `json:"-"`
	Password      string   `json:"-" db:"userPassword"`
	Salt          zero.Int `json:"-" db:"userPasswordSalt"`
	TimezoneId    int64    `json:"timezoneId" db:"timezoneId"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// State fields
	ZuluCreatedAt string `json:"createdAt"`
	ZuluUpdatedAt string `json:"updatedAt"`
	State         string `json:"status"`
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