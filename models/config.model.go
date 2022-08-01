package models

type Config struct {
	Cron         string   `json:"cron"`
	Language     string   `json:"language"`
	Script       string   `json:"script"`
	Independant  bool     `json:"independant"`
	Dependencies []string `json:"dependencies"`
}
