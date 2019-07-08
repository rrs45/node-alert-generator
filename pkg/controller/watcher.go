package controller

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	coreInformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

//AlertGeneratorController struct for encapsulating generic Informer methods and Node informer
type AlertGeneratorController struct {
	informerFactory     informers.SharedInformerFactory
	nodeInformer        coreInformers.NodeInformer
	filterCh       		chan<- *v1.Node
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *AlertGeneratorController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so far
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync informer cache")
	}
	return nil
}

func (c *AlertGeneratorController) nodeAdd(obj interface{}) {
	node := obj.(*v1.Node)
	log.Infof("Watcher - Received node add event for %s in watcher.go ", node.Name)
}

func (c *AlertGeneratorController) nodeUpdate(oldN, newN interface{}) {
	newNode := newN.(*v1.Node)
	//Exclude cordon'd nodes
	if !newNode.Spec.Unschedulable {
		c.filterCh <- newNode
	}	
	//Include both cordon'd and uncordon'd nodes
	c.filterCh <- newNode
}

func (c *AlertGeneratorController) nodeDelete(obj interface{}) {
	node := obj.(*v1.Node)
	log.Infof("Watcher - Received node delete event for %s in watcher.go", node.Namespace)
}

//NewAlertGeneratorController creates a initializes AlertGeneratorController struct
//and adds event handler functions
func NewAlertGeneratorController(informerFactory informers.SharedInformerFactory, filterCh chan<- *v1.Node ) *AlertGeneratorController {
	nodeInf := informerFactory.Core().V1().Nodes()

	c := &AlertGeneratorController{
		informerFactory:     informerFactory,
		nodeInformer:        nodeInf,
		filterCh:             filterCh,
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

//Start starts the controller
func Start(clientset *kubernetes.Clientset, filterCh chan<- *v1.Node) {

	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	//log.SetNoLock()
	log.Info("Watcher - Creating informer factory for alert-generator ")
	//Create shared cache informer which resync's every 24hrs
	factory := informers.NewFilteredSharedInformerFactory(clientset, time.Hour*24, "", func(opt *metav1.ListOptions) { opt.LabelSelector = "box.com/pool in (generic, calico)" })
	controller := NewAlertGeneratorController(factory, filterCh)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		log.Error("Watcher - Could not run controller :", err)
	}
	select {}
}
