package config

import (
	"github.com/spf13/viper"
	"runtime"
	"strings"
	"sync"
)

type Config struct {
	Mail Mail `mapstructure:"mail" yaml:"mail"`
}

type Mail struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

var (
	conf Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		configFilePath := GetCurrentDirPath() + "/"
		viper.SetConfigType("yml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configFilePath)
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
		err = viper.Unmarshal(&conf)
		if err != nil {
			panic(err)
		}
	})
	return &conf
}

func GetCurrentDirPath() string {
	_, filename, _, _ := runtime.Caller(1)
	aPath := strings.Split(filename, "/")
	return strings.Join(aPath[:len(aPath)-1], "/")
}
