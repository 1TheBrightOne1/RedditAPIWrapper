package config

import "flag"

type Config struct {
	HomePath string
}

var GlobalConfig Config

func InitConfig() {
	flag.StringVar(&GlobalConfig.HomePath, "homepath", "/var/stonks", "where persisted data will be stored")
	flag.Parse()
}
