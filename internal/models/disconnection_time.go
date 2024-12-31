package models

import "time"

type DisconnectionTime struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
