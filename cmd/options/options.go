package options

import (
	"flag"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

type AlertGeneratorOptions struct {
	ServerAddress       string
	ServerPort          string
	ApiServerHost       string
	NoLabel             bool
	UpdateInterval      string
	LogFile             string
	Namespace           string
	AlertConfigMap      string
	AlertIgnoreInterval string
}

func NewAlertGeneratorOptions() *AlertGeneratorOptions {
	return &AlertGeneratorOptions{}
}

func (ago *AlertGeneratorOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&ago.ServerAddress, "address", "127.0.0.1",
		"Address to bind the alert generator server.")
	fs.StringVar(&ago.ServerPort, "port", "8080", "Port to bind the alert generator server for /healthz endpoint")
	fs.StringVar(&ago.ApiServerHost, "apiserver-host", "", "Custom hostname used to connect to Kubernetes ApiServer")
	fs.BoolVar(&ago.NoLabel, "nolabel", true, "Dont set labels")
	fs.StringVar(&ago.LogFile, "log-file", "/var/log/service/node-alert-generator.log", "Log file to store all logs")
	fs.StringVar(&ago.UpdateInterval, "alert-update-interval", "60s", "Interval in seconds at which configmap will be updated")
	fs.StringVar(&ago.AlertConfigMap, "alert-config-map", "npd-alerts", "Name of config map to store alerts")
	fs.StringVar(&ago.Namespace, "namespace", "node-alert-generator", "Namespace where config map will be")
	fs.StringVar(&ago.AlertIgnoreInterval, "alert-ignore-interval", "25h", "Ignore alerts older than this interval")
}

func (ago *AlertGeneratorOptions) ValidOrDie() {
	_, err := time.ParseDuration(ago.UpdateInterval)
	if err != nil {
		log.Error("Options - Incorrect alert-update-interval, sample format: 10s or 1m or 1h; ", err)
	}
	dir, _ := path.Split(ago.LogFile)
	_, err1 := os.Stat(dir)
	if err1 != nil {
		log.Errorf("Options - Directory does not exist: %s ", dir)
	}
	_, err2 := time.ParseDuration(ago.AlertIgnoreInterval)
	if err2 != nil {
		log.Error("Options - Incorrect alert-ignore-interval, sample format: 10s or 1m or 1h; ", err)
	}
	if err != nil || err1 != nil || err2 != nil {
		log.Panic("Incorrect options")
	}
}
