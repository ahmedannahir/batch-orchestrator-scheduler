package models

type Config struct {
	Cron         string   `json:"cron"`
	Language     string   `json:"language"`
	Script       string   `json:"script"`
	Dependencies []string `json:"dependencies"`
}
