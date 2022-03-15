package models

import "time"

type RequestBalance struct {
	Id string `json:"id"`
}

type ResponseBalance struct {
	UserId  string  `json:"id"`
	Balance float64 `json:"balance"`
}

type RequestChangeBalance struct {
	Operation string  `json:"operation"`
	UserId    string  `json:"id"`
	Sum       float64 `json:"sum"`
}

type RequestTransfer struct {
	SenderId   string  `json:"sender_id"`
	ReceiverId string  `json:"receiver_id"`
	Sum        float64 `json:"sum"`
}

type RequestLogs struct {
	Id             string `json:"id"`
	CountOperation int    `json:"count"`
}

type ResponsetLogs struct {
	Logs []Log `json:"logs"`
}

type Log struct {
	Id          string    `json:"id"`
	UserId      string    `json:"userId"`
	Description string    `json:"description"`
	Date        time.Time `json:"date`
}
