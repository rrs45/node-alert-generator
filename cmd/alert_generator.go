package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"

	//_ "runtime/pprof"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/box-autoremediation/cmd/options"
	"github.com/box-autoremediation/pkg/controller"
	"github.com/box-autoremediation/pkg/controller/types"
)

func initClient(ago *options.AlertGeneratorOptions) (*kubernetes.Clientset, error) {

	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
		return kubernetes.NewForConfig(config)
	} else {
		kubeConfig, err := restclient.InClusterConfig()
		if err != nil {
			panic(err)
		}
		if ago.ApiServerHost != "" {
			kubeConfig.Host = ago.ApiServerHost
		}
		return kubernetes.NewForConfig(kubeConfig)
	}
}

func main() {
	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	/*flog := log.WithFields(log.Fields{
		"file": "cmd/alert_generator.go",
	}) */
	ago := options.NewAlertGeneratorOptions()
	ago.AddFlags(flag.CommandLine)
	flag.Parse()
	//Instantiate the http server
	addr := ago.ServerAddress + ":" + ago.ServerPort
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	var wg sync.WaitGroup
	wg.Add(3)

	// Create an rest client not targeting specific API version
	log.Info("Calling initClient for alert-generator")
	clientset, err := initClient(ago)
	if err != nil {
		panic(err)
	}
	log.Info("Successfully generated k8 client for alert-generator")
	alertch := make(chan types.Alert)
	labelch := make(chan *v1.Node)

	//GOroutine to run the controller
	log.Info("Starting controller for alert-generator")
	go func() {
		controller.Do(clientset, ago.NoLabel, alertch, labelch)
		log.Info("Controller stopped for alert-generator")
		wg.Done()
	}()

	log.Info("Starting labeller for alert-generator")
	go controller.LabelNode(clientset, &wg, labelch)

	log.Info("Starting updater for alert-generator")
	go func() {
		controller.Update(clientset, ago.UpdateInterval, alertch)
		wg.Done()
	}()

	//Goroutine for serve healthz endpoint
	log.Info("Starting HTTP server for alert-generator")
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	wg.Wait()
}
