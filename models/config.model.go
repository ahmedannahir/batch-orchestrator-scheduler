package models

type Config struct {
	Cron           string   `json:"cron"`
	Script         string   `json:"script"`
	PrevBatchInput bool     `json:"prevBatchInput"`
	Independant    bool     `json:"independant"`
	Args           []string `json:"args"`
}
