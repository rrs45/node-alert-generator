package options

import (
	"flag"
	"time"
	 log "github.com/sirupsen/logrus"
)

type AlertGeneratorOptions struct {
	ServerAddress  string
	ServerPort     string
	ApiServerHost  string
	NoLabel        bool
	UpdateInterval string
}

func NewAlertGeneratorOptions() *AlertGeneratorOptions {
	return &AlertGeneratorOptions{}
}

func (ago *AlertGeneratorOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&ago.ServerAddress, "address", "127.0.0.1",
		"Address to bind the alert generator server.")
	fs.StringVar(&ago.ServerPort, "port", "8080", "Port to bind the alert generator server")
	fs.StringVar(&ago.ApiServerHost, "apiserver-host", "", "Custom hostname used to connect to Kubernetes ApiServer")
	fs.BoolVar(&ago.NoLabel, "nolabel", true, "Dont set labels")
	fs.StringVar(&ago.UpdateInterval, "interval", "60", "Interval in seconds at which configmap will be updated")
}

func (ago *AlertGeneratorOptions) ValidOrDie() {
	_, err := time.ParseDuration(ago.UpdateInterval)
	if err != nil {
		log.Panic("Updater - Incorrect interval, sample format: 10s or 1m or 1h; ", err)
	}
}
