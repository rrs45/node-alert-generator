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

	"github.com/box-autoremediation/pkg/controller/types"
)

//AlertGeneratorController struct for encapsulating generic Informer methods and Node informer
type AlertGeneratorController struct {
	informerFactory informers.SharedInformerFactory
	nodeInformer    coreInformers.NodeInformer
	alertch         chan<- types.Alert
	labelch         chan<- *v1.Node
	nolabel         bool
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

func checkLabels(labels map[string]string) string {
	for k, v := range labels {
		if k == "maintenance.box.com/source" {
			if v != "npd" {
				return "non_npd_maint"
			} else {
				return "npd_maint"
			}
		}
	}
	return "no_maint"
}

func checkConditions(conditions []v1.NodeCondition, node string) ([]types.Alert, bool) {
	var buf []types.Alert
	nodeReady := false
	var item types.Alert
	for _, condition := range conditions {
		if condition.Type[:4] == "NPD-" && condition.Status == "True" {
			item.Timestamp = condition.LastHeartbeatTime.Time
			item.Node = node
			item.Condition = condition.Type
			item.Action = ""
			item.Params = condition.Message
			buf = append(buf, item)
		} else if condition.Type == "Ready" && condition.Status == "True" {
			nodeReady = true
		}
	}
	return buf, nodeReady
}

func (c *AlertGeneratorController) nodeUpdate(oldN, newN interface{}) {
	oldNode := oldN.(*v1.Node)
	//generateAlert(oldNode, c.nolabel, c.alertch, c.labelch)
	var labeled bool
	//Check if node is cordon'd & has maintenance labels
	if !oldNode.Spec.Unschedulable {
		maint := checkLabels(oldNode.GetLabels())
		if maint == "non_npd_maint" {
			log.Info("Watcher - Node under maintenance by different source")
			return
		} else if maint == "npd_maint" {
			labeled = true
		}
		buf, nodeReady := checkConditions(oldNode.Status.Conditions, oldNode.Name)
		//log.Info(buf)
		if nodeReady && buf != nil {
			log.Debug("Watcher - Found issue on ", oldNode.Name, " in watcher.go")
			for _, a := range buf {
				c.alertch <- a
			}
			if !labeled && !c.nolabel {
				c.labelch <- oldNode
			}
		}
	}
}

func (c *AlertGeneratorController) nodeDelete(obj interface{}) {
	node := obj.(*v1.Node)
	log.Infof("Watcher - Received node delete event for %s in watcher.go", node.Namespace)
}

//NewAlertGeneratorController creates a initializes AlertGeneratorController struct
//and adds event handler functions
func NewAlertGeneratorController(informerFactory informers.SharedInformerFactory, nolabel bool, alertch chan<- types.Alert, labelch chan<- *v1.Node) *AlertGeneratorController {
	nodeInf := informerFactory.Core().V1().Nodes()

	c := &AlertGeneratorController{
		informerFactory: informerFactory,
		nodeInformer:    nodeInf,
		alertch:         alertch,
		labelch:         labelch,
		nolabel:         nolabel,
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
func Start(clientset *kubernetes.Clientset, nolabel bool, alertch chan<- types.Alert, labelch chan<- *v1.Node) {
	//Get current directory

	//Set logrus
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	//log.SetNoLock()
	log.Info("Watcher - Creating informer factory for alert-generator ")
	//Create shared cache informer which resync's every 24hrs
	factory := informers.NewFilteredSharedInformerFactory(clientset, time.Hour*24, "", func(opt *metav1.ListOptions) { opt.LabelSelector = "box.com/pool in (generic, calico)" })
	controller := NewAlertGeneratorController(factory, nolabel, alertch, labelch)
	stop := make(chan struct{})
	defer close(stop)
	err := controller.Run(stop)
	if err != nil {
		log.Error("Watcher - Could not run controller :", err)
	}
	select {}
}
