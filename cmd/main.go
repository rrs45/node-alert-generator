package main

import (
	"context"
	_ "fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/box-autoremediation/alertgenerator"
)

const Dir = "box-autoremediation"

func main() {

	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	flog := log.WithFields(log.Fields{
		"file": Dir + "/cmd/main.go",
	})

	//Instantiate the http server
	server := &http.Server{Addr: ":8080", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})}

	var wg sync.WaitGroup
	wg.Add(1)

	// Bootstrap k8s configuration from local     Kubernetes config file
	kubeconfig := new(string)
	*kubeconfig = ""
	masterUrl := new(string)
	*masterUrl = "localhost:8080"
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
		*kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		flog.Info("Using ", *kubeconfig)
		*masterUrl = ""
	}
	config, err := clientcmd.BuildConfigFromFlags(*masterUrl, *kubeconfig)
	if err != nil {
		flog.WithFields(log.Fields{
			"function": "clientcmd.BuildConfigFromFlags",
		}).Error(err)
	}
	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		flog.WithFields(log.Fields{
			"function": "kubernetes.NewForConfig",
		}).Error(err)
	}

	//GOroutine to run the controller
	go func() {
		alertgenerator.Do(clientset)
		if err := server.Shutdown(context.Background()); err != nil {
			flog.WithFields(log.Fields{
				"function": "server.Shutdown(context.Background())",
			}).Error(err)
		}
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
