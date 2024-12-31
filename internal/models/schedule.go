package models

import "time"

type Schedule struct {
	FetchTime time.Time `json:"fetchTime"`
	Queues    []Queue   `json:"queues"`
}
