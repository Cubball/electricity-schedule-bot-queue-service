package models

type Queue struct {
	Number             string              `json:"number"`
	DisconnectionTimes []DisconnectionTime `json:"disconnectionTimes"`
}
