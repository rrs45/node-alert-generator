package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	coreInformers "k8s.io/client-go/informers/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

//Struct for encapsulating generic Informer methods and Node informer
type AlertGeneratorController struct {
	informerFactory informers.SharedInformerFactory
	nodeInformer	coreInformers.NodeInformer
} 

//Defines format for alerts detected from NPD
type Alert struct {
	empty     bool
	Timestamp time.Time
	node      string
	condition v1.NodeConditionType
	action    string
	params    string
}

//For future use
func mapToString(m map[string]string) string {
	b := new(bytes.Buffer)
	for k, v := range m {
		if k != "kubernetes.io/hostname" {
			fmt.Fprintf(b, "%s=\"%s\"\n", k, v)
		}
	}
	return b.String()
}

/*Filters & prints based on :
1. Nodes are Ready and Uncordon'd
2. NPD conditions which are True
*/
func PrintAlert(n *v1.Node) {
	a := n.Status.Conditions
	x := new(Alert)
	node_ready := false
	for _, i := range a {
		x.empty = true
		if i.Type[:4] == "NPD-"  && i.Status == "True" {
			x.empty = false
			x.Timestamp= i.LastHeartbeatTime.Time
			x.node= n.Name
			x.condition= i.Type
			x.action= ""
			x.params= i.Message
		} else if i.Type == "Ready" && i.Status == "True"{
			node_ready = true
		}
		
		if node_ready && !x.empty {
			//fmt.Println(time.Now())
			fmt.Println(x)
		}
	}
}


// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *AlertGeneratorController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *AlertGeneratorController) nodeAdd(obj interface{}) {
	node := obj.(*v1.Node)
	log.Println("NODE CREATED: ", node.Name)
}

func (c *AlertGeneratorController) nodeUpdate(old, new interface{}) {
	oldNode := old.(*v1.Node)
	PrintAlert(oldNode)
}

func (c *AlertGeneratorController) nodeDelete(obj interface{}) {
	node := obj.(*v1.Node)
	log.Println("NODE DELETED: %s/%s", node.Namespace, node.Name)
}

// NewAlertGeneratorController creates a new AlertGeneratorController
func NewAlertGeneratorController(informerFactory informers.SharedInformerFactory) *AlertGeneratorController {
	nodeInf := informerFactory.Core().V1().Nodes()

	c := &AlertGeneratorController{
		informerFactory: informerFactory,
		nodeInformer: nodeInf,
	}
	nodeInf.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.nodeAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.nodeUpdate,
			// Called on resource deletion.
			DeleteFunc: c.nodeDelete,
		},
	)
	return c
}

func main() {

	// Bootstrap k8s configuration from local     Kubernetes config file
	kubeconfig := new(string)
	*kubeconfig = ""
	masterUrl := new(string)
	*masterUrl = "localhost:8080"
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
		*kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		log.Println("Using kubeconfig file: ", *kubeconfig)
		*masterUrl = ""
	}
	config, err := clientcmd.BuildConfigFromFlags(*masterUrl, *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	
	//Create shared cache informer which resync's every 24hrs
	factory := informers. NewFilteredSharedInformerFactory(clientset, time.Hour*24,"",func(opt *metav1.ListOptions) {opt.LabelSelector = "box.com/pool in (generic, calico)" })
	controller := NewAlertGeneratorController(factory)
	stop := make(chan struct{})
	defer close(stop)
	err = controller.Run(stop)
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
