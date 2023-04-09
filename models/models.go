package models

import "time"

type Todo struct {
	ID          string    `json:id`
	UserID      string    `json:userID`
	Title       string    `json:title`
	Description string    `json:description`
	Deadline    time.Time `json:deadline`
	Completed   bool      `json:completed`
}
