package models

import "time"

type ScheduleUpdate struct {
	FetchTime     time.Time `json:"fetchTime"`
	UpdatedQueues []Queue   `json:"updatedQueues"`
}
