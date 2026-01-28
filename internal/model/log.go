package model

import "time"

type MikrotikLog struct {
	Time    string `json:"time"`
	Topics  string `json:"topics"`
	Message string `json:"message"`
}

type LogStreamData struct {
	Timestamp time.Time   `json:"timestamp"`
	Log       MikrotikLog `json:"log"`
}
