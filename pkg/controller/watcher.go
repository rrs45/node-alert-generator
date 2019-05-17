package controller

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	coreInformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/box-autoremediation/pkg/controller/types"
)

//Struct for encapsulating generic Informer methods and Node informer
type AlertGeneratorController struct {
	informerFactory informers.SharedInformerFactory
	nodeInformer    coreInformers.NodeInformer
	alertch         chan<- types.Alert
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *AlertGeneratorController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so far
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *AlertGeneratorController) nodeAdd(obj interface{}) {
	node := obj.(*v1.Node)
	log.Info("NODE CREATED: ", node.Name)
}

func (c *AlertGeneratorController) nodeUpdate(old, new interface{}) {
	oldNode := old.(*v1.Node)
	var x types.Alert
	//Check if node is cordon'd & has maintenance labels
	if !oldNode.Spec.Unschedulable {
		for k, _ := range oldNode.GetLabels() {
			if strings.Contains(k, "maintenance.box.com") {
				goto Exit
			}
		}
		node_ready := false
		for _, i := range oldNode.Status.Conditions {
			x.Empty = true
			if i.Type[:4] == "NPD-" && i.Status == "True" {
				x.Empty = false
				x.Timestamp = i.LastHeartbeatTime.Time
				x.Node = oldNode.Name
				x.Condition = i.Type
				x.Action = ""
				x.Params = i.Message
			} else if i.Type == "Ready" && i.Status == "True" {
				node_ready = true
			}
		}

		if node_ready && !x.Empty {
			c.alertch <- x
		}
	}
Exit:
}

func (c *AlertGeneratorController) nodeDelete(obj interface{}) {
	node := obj.(*v1.Node)
	log.Info("NODE DELETED: %s/%s", node.Namespace, node.Name)
}

// NewAlertGeneratorController creates a new AlertGeneratorController
func NewAlertGeneratorController(informerFactory informers.SharedInformerFactory, ch chan<- types.Alert) *AlertGeneratorController {
	nodeInf := informerFactory.Core().V1().Nodes()

	c := &AlertGeneratorController{
		informerFactory: informerFactory,
		nodeInformer:    nodeInf,
		alertch:         ch,
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

func Do(clientset *kubernetes.Clientset, ch chan<- types.Alert) {
	//Get current directory

	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	//log.SetNoLock()
	flog := log.WithFields(log.Fields{
		"file": "pkg/controller/watcher.go",
	})

	//Create shared cache informer which resync's every 24hrs
	factory := informers.NewFilteredSharedInformerFactory(clientset, time.Hour*24, "", func(opt *metav1.ListOptions) { opt.LabelSelector = "box.com/pool in (generic, calico)" })
	controller := NewAlertGeneratorController(factory, ch)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		flog.WithFields(log.Fields{
			"function": "controller.Run(stop)",
		}).Error(err)
	}
	select {}
}
