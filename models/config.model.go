package models

type Config struct {
	Cron         string   `json:"cron"`
	Language     string   `json:"language"`
	Dependencies []string `json:"dependencies"`
}
