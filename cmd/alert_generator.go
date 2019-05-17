package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"

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
	flog := log.WithFields(log.Fields{
		"file": "cmd/alert_generator.go",
	})
	ago := options.NewAlertGeneratorOptions()
	//ago.AddFlags(flag.NewFlagSet(os.Args[0], flag.ExitOnError))
	ago.AddFlags(flag.CommandLine)
	flag.Parse()
	//Instantiate the http server
	addr := ago.ServerAddress + ":" + ago.ServerPort
	server := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})}

	var wg sync.WaitGroup
	wg.Add(2)

	// Create an rest client not targeting specific API version
	clientset, err := initClient(ago)
	if err != nil {
		panic(err)
	}
	alertch := make(chan types.Alert)

	//GOroutine to run the controller
	go func() {
		controller.Do(clientset, alertch)
		if err := server.Shutdown(context.Background()); err != nil {
			flog.WithFields(log.Fields{
				"function": "server.Shutdown(context.Background())",
			}).Error(err)
		}
		wg.Done()
	}()

	go func() {
		controller.Update(clientset, alertch)
		wg.Done()
	}()

	//Goroutine for serve healthz endpoint
	go func() {
		if err := server.ListenAndServe(); err != nil {
			flog.WithFields(log.Fields{
				"function": "server.ListenAndServe",
			}).Error(err)
		}
	}()

	wg.Wait()
}
