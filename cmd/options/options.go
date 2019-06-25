package options

import (
	"flag"
	"os"
	"time"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//NewConfigFromFile parses config file  
func NewConfigFromFile(configFile string) (*viper.Viper, error) {
	//dir, file := filepath.Split(configFile)
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(configFile)
	//v.AddConfigPath(dir)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

//Config defines configuration parameters
type Config struct {
	File string
	KubeAPIURL string
}

//GetConfig returna new config file
func GetConfig() *Config {
	return &Config{}
}

//AddFlags takes config file input
func (c *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.File, "file", "/home/rajsingh/go/src/github.com/box-autoremediation/config/config.toml", "Configuration file path")
	fs.StringVar(&c.KubeAPIURL, "apiserver-override", "", "URL of the kubernetes api server")
}

//ValidOrDie validates some of the config parameters
func ValidOrDie(ago *viper.Viper) {
	log.Infof("%+v",ago.AllSettings())
	_, err := time.ParseDuration(ago.GetString("config_map.frequency"))
	if err != nil {
		log.Errorf("Options - Incorrect config_map.frequency: %v ", err)
	}
	dir, _ := filepath.Split(ago.GetString("log_file"))
	_, err1 := os.Stat(dir)
	if err1 != nil {
		log.Errorf("Options - Directory does not exist: %v ", err1)
	} 
	_, err2 := time.ParseDuration(ago.GetString("condition.options.interval"))
	if err2 != nil {
		log.Errorf("Options - conditions_filter.interval: %v ", err)
	}
	if err != nil || err1 != nil || err2 != nil {
		log.Panic("Incorrect options")
	}
}
