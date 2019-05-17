package options

import (
	"flag"
)

type AlertGeneratorOptions struct {
	ServerAddress string
	ServerPort    string
	ApiServerHost string
	NoLabel       bool
}

func NewAlertGeneratorOptions() *AlertGeneratorOptions {
	return &AlertGeneratorOptions{}
}

func (ago *AlertGeneratorOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&ago.ServerAddress, "address", "127.0.0.1",
		"Address to bind the alert generator server.")
	fs.StringVar(&ago.ServerPort, "port", "8080", "Port to bind the alert generator server")
	fs.StringVar(&ago.ApiServerHost, "apiserver-host", "", "Custom hostname used to connect to Kubernetes ApiServer")
	fs.BoolVar(&ago.NoLabel, "nolabel", false, "Dont set labels")
}
