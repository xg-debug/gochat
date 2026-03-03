package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)


type MainConfig struct {
	AppName string `yaml:"appName"`
	Host    string `yaml:"host"`
	Port    int	   `yaml:"port"`
}

type MysqlConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"databaseName"`
}

type RedisConfig struct {
	Host 	 string     `yaml:"host"`
	Port 	 int    	`yaml:"port"`
	Password string 	`yaml:"password"`
	Db       int    	`yaml:"db"`
}

type KafkaConfig struct {
	MessageMode string        `toml:"messageMode"`
	HostPort    string        `toml:"hostPort"`
	LoginTopic  string        `toml:"loginTopic"`
	LogoutTopic string        `toml:"logoutTopic"`
	ChatTopic   string        `toml:"chatTopic"`
	Partition   int           `toml:"partition"`
	Timeout     time.Duration `toml:"timeout"`
}

type LogConfig struct {
	LogPath string `yaml:"logPath"`
}

type JwtConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expireHours"`
}

type Config struct {
	MainConfig  `yaml:"MainConfig"`
	MysqlConfig `yaml:"MysqlConfig"`
	RedisConfig `yaml:"RedisConfig"`
	KafkaConfig `yaml:"KafkaConfig"`
	LogConfig   `yaml:"LogConfig"`
	JwtConfig   `yaml:"JwtConfig"`
}

var config *Config

func LoadConfig() error {
	viper.SetConfigFile("../configs/config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("读取配置失败:", err)
		return err
	}

	if config == nil {
		config = new(Config)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("解析配置失败:", err)
		return err
	}

	// 本地部署
	// if _, err := toml.DecodeFile("F:\\go\\kama-chat-server\\configs\\config_local.toml", config); err != nil {
	// 	log.Fatal(err.Error())
	// 	return err
	// }
	// Ubuntu22.04云服务器部署
	// if _, err := toml.DecodeFile("/root/project/gochat/configs/config_local.toml", config); err != nil {
	// 	log.Fatal(err.Error())
	// 	return err
	// }
	return nil
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
