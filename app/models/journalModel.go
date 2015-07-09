package models

import (
	"time"

	"github.com/coopernurse/gorp"
)

type JournalEntry struct {
	// Identity fields
	Id int64 `json:"id"`
	// Standard fields A-Z
	Users_id int64     `db:"users_id" json:"userId"`
	Date     time.Time `db:"date" json:"date"`
	Body     string    `db:"body" json:"body"`
	// State fields
	State string `db:"state" json:"-"`
}

func (i *JournalEntry) PreInsert(s gorp.SqlExecutor) error {
	if i.State == "" {
		i.State = "published"
	}
	return nil
}
