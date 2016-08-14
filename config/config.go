package config

import "github.com/klarna/eremetic/database"

type Config struct {
	Version   string
	BuildDate string
	Database  database.TaskDB
}
