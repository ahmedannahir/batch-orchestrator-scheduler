package models

type Config struct {
	JobName        string   `json:"jobName"`
	JobDesc        string   `json:"jobDesc"`
	Cron           string   `json:"cron"`
	Script         string   `json:"script"`
	PrevBatchInput bool     `json:"prevBatchInput"`
	Independant    bool     `json:"independant"`
	Args           []string `json:"args"`
}
