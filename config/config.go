package config

import "flag"

type Config struct {
	HomePath string
}

var GlobalConfig Config

func InitConfig() {
	flag.StringVar(&GlobalConfig.HomePath, "homepath", "E:/Programming Projects/go/src/github.com/1TheBrightOne1/RedditAPIWrapper", "where persisted data will be stored")
	flag.Parse()
}
