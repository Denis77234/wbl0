package postgres

import (
	"fmt"
)

type Config struct {
	Login    string
	Password string
	HostName string
	Port     string
	Database string
	Ssl      string
}

func (c Config) connString() string {
	return fmt.Sprintf("host=%s user=%s password=%s  dbname=%s port=%v  sslmode=%s", c.HostName, c.Login, c.Password, c.Database, c.Port, c.Ssl)
}
