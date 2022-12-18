package config

type MySQL struct {
	Addr     string `json:"addr"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}
