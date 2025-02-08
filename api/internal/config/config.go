package config

import (
    "flag"
)

type Config struct {
    Host string
    Port string
	DBHost       string // Хост базы данных
    DBPort       string // Порт базы данных
    DBUser       string // Пользователь базы данных
    DBPassword   string // Пароль базы данных
    DBName       string // Имя базы данных
}

func LoadConfig() *Config {
    cfg := &Config{}
    flag.StringVar(&cfg.Host, "host", "0.0.0.0", "Server host")
    flag.StringVar(&cfg.Port, "port", "8080", "Server port")
    flag.StringVar(&cfg.DBHost, "dbhost", "my-postgres", "Database host")
    flag.StringVar(&cfg.DBPort, "dbport", "5432", "Database port")
    flag.StringVar(&cfg.DBUser, "dbuser", "postgres", "Database user")
    flag.StringVar(&cfg.DBPassword, "dbpassword", "password", "Database password")
    flag.StringVar(&cfg.DBName, "dbname", "mydb", "Database name")
    flag.Parse()
    return cfg
}