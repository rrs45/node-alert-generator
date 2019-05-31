package main

import (
	"context"
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

func startHTTPServer(addr string, port string) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: addr + ":" + port, Handler: mux}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		log.Info("Starting HTTP server for alert-generator")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Could not start http server: %s", err)
		}
	}()
	return srv
}

func main() {
	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	//Parse command line options
	ago := options.NewAlertGeneratorOptions()
	ago.AddFlags(flag.CommandLine)
	flag.Parse()
	ago.ValidOrDie()
	f, _ := os.OpenFile(ago.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	log.SetOutput(f)

	srv := startHTTPServer(ago.ServerAddress, ago.ServerPort)

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

	go func() {
		log.Info("Starting controller for alert-generator")
		controller.Start(clientset, ago.NoLabel, ago.AlertIgnoreInterval, alertch, labelch)
		log.Info("Stopping controller for alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	go func() {
		log.Info("Starting labeller for alert-generator")
		controller.LabelNode(clientset, labelch)
		log.Info("Stopping labeller for alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	go func() {
		log.Info("Starting updater for alert-generator")
		controller.Update(clientset, ago.Namespace, ago.AlertConfigMap, ago.UpdateInterval, alertch)
		log.Info("Stopping updater for alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
