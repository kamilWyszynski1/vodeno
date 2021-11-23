package client

import "time"

// Entry represent client entry.
type Entry struct {
	ID         int       `json:"id" db:"id"`
	Email      string    `json:"email" db:"email" validate:"required,email"`
	Title      string    `json:"title" db:"title" validate:"required"`
	Content    string    `json:"content" db:"content" validate:"required"`
	MailingID  int       `json:"mailing_id" db:"mailing_id" validate:"required"`
	InsertTime time.Time `json:"insert_time" db:"insert_time" validate:"required"`
}
