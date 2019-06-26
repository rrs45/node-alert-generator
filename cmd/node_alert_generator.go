package main

import (
	"context"
	"flag"
	"io"
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

func initClient(kubeAPI string) (*kubernetes.Clientset, error) {

	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
		return kubernetes.NewForConfig(config)
	} 
	kubeConfig, err := restclient.InClusterConfig()
	
	if err != nil {
		panic(err)
	}
	if kubeAPI != "" {
		kubeConfig.Host = kubeAPI
		return kubernetes.NewForConfig(kubeConfig)
	}
	return kubernetes.NewForConfig(kubeConfig)
}

func startHTTPServer(addr string, port string) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: addr + ":" + port, Handler: mux}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		log.Info("Starting HTTP server for node-alert-generator")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Could not start http server: %s", err)
		}
	}()
	return srv
}

func main() {
	//Parse command line options
	conf := options.GetConfig()
	conf.AddFlags(flag.CommandLine)
	flag.Parse()
	nago, err := options.NewConfigFromFile(conf.File)
	if err!= nil {
		log.Fatalf("Cannot parse config file: %v", err)
	}
	options.ValidOrDie(nago)
	logFile, _ := os.OpenFile(nago.GetString("log_file"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	defer logFile.Close()
	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	srv := startHTTPServer(nago.GetString("healthcheck.server"), nago.GetString("healthcheck.port"))

	var wg sync.WaitGroup
	wg.Add(4)

	// Create an rest client not targeting specific API version
	log.Info("Calling initClient for node-alert-generator")
	clientset, err := initClient(conf.KubeAPIURL)
	if err != nil {
		panic(err)
	}
	log.Info("Successfully generated k8 client for node-alert-generator")
	filterCh := make(chan *v1.Node)
	alertch := make(chan types.Alert)
	labelch := make(chan *v1.Node)

	//Watcher
	go func() {
		log.Info("Starting controller for node-alert-generator")
		controller.Start(clientset, nago.GetBool("node_status.cordoned"), filterCh)
		log.Info("Stopping controller for node-alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	//Filter
	go func() {
		log.Info("Starting node filter for node-alert-generator")
		controller.Filter(filterCh, labelch, alertch, nago)
		log.Info("Stopping labeller for node-alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	//Label
	go func() {
		log.Info("Starting labeller for node-alert-generator")
		controller.LabelNode(clientset, labelch, nago.GetString("NodeLabel"))
		log.Info("Stopping labeller for node-alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	go func() {
		log.Info("Starting updater for node-alert-generator")
		controller.Update(clientset, nago.GetString("namespace"), nago.GetString("config_map.name"), nago.GetString("config_map.frequency"), alertch)
		log.Info("Stopping updater for node-alert-generator")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not stop http server: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
