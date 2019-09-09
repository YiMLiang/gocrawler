package config

import (
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

type Config struct {
	LogPath  string
	LogLevel string
}

var (
	Conf *Config
)

func LoadConf(configType, fileName string) (err error) {

	conf, err := config.NewConfig(configType, fileName)
	if err != nil {
		logs.Error("LoadConf failed err = ,", err)
		return
	}
	Conf = &Config{}
	Conf.LogLevel = conf.String("logs::log_level")
	if len(Conf.LogLevel) == 0 {
		Conf.LogLevel = "debug"
	}
	Conf.LogPath = conf.String("logs::log_path")
	if len(Conf.LogPath) == 0 {
		Conf.LogPath = "./logs"
	}
	return err
}
