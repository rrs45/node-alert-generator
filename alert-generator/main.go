package main

import (
        "fmt"
        "time"
	"os"
	"path/filepath"
	"log"
	"bytes"
        "k8s.io/client-go/informers"
	//"k8s.io/apimachinery/pkg/labels"
        coreinformers "k8s.io/client-go/informers/core/v1"
        "k8s.io/client-go/kubernetes"
        //"k8s.io/client-go/pkg/api/v1"
        "k8s.io/api/core/v1"
        "k8s.io/client-go/tools/cache"
        "k8s.io/client-go/tools/clientcmd"
)

func MapToString(m map[string]string) string {
    b := new(bytes.Buffer)
    for k, v := range m {
        if k != "kubernetes.io/hostname" {
            fmt.Fprintf(b, "%s=\"%s\"\n", k , v)
        }
    }
    return b.String()
}

func NodeCondPrint(a []v1.NodeCondition) string {
    b := new(bytes.Buffer)
    for _, i := range a {
        if i.Type == "Ready"  &&  i.Status == "True"{
            //fmt.Fprintf(b, "%v\n", i
            fmt.Fprintf(b,"Ready")
        } else if i.Type == "Ready"  &&  i.Status != "True"{
            fmt.Fprintf(b,"NotReady")
        }
    }
    return b.String()
}

// NodeLoggingController logs the name and namespace of nodes that are added,
// deleted, or updated
type NodeLoggingController struct {
        informerFactory informers.SharedInformerFactory
        nodeInformer     coreinformers.NodeInformer
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *NodeLoggingController) Run(stopCh chan struct{}) error {
        // Starts all the shared informers that have been created by the factory so
        // far.
        c.informerFactory.Start(stopCh)
        // wait for the initial synchronization of the local cache.
        if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
                return fmt.Errorf("Failed to sync")
        }
        return nil
}

func (c *NodeLoggingController) nodeAdd(obj interface{}) {
        node := obj.(*v1.Node)
        log.Println("NODE CREATED: %s/%s", node.Namespace, node.Name)
}

func (c *NodeLoggingController) nodeUpdate(old, new interface{}) {
        oldNode := old.(*v1.Node)
        //newNode := new.(*v1.Node)
        log.Println(
                "NODE UPDATED: ",
                 oldNode.Name, NodeCondPrint(oldNode.Status.Conditions), MapToString(oldNode.Labels),
        )
}

func (c *NodeLoggingController) nodeDelete(obj interface{}) {
        node := obj.(*v1.Node)
        log.Println("NODE DELETED: %s/%s", node.Namespace, node.Name)
}

// NewNodeLoggingController creates a NodeLoggingController
func NewNodeLoggingController(informerFactory informers.SharedInformerFactory) *NodeLoggingController {
        //labelSelector := labels.Set(map[string]string{"box.com/pool": "master"}).AsSelector()
        //nodeInformer := informerFactory.Core().V1().Nodes.Lister().List(labelSelector.String())
        nodeInformer := informerFactory.Core().V1().Nodes()

        c := &NodeLoggingController{
                informerFactory: informerFactory,
                nodeInformer:     nodeInformer,
        }
        nodeInformer.Informer().AddEventHandler(
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
    if _,err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err == nil {
        *kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
        log.Println("Using kubeconfig file: ", *kubeconfig)
        *masterUrl = ""
    }
    config, err := clientcmd.BuildConfigFromFlags(*masterUrl,*kubeconfig)
    if err != nil {
               log.Fatal(err)
           }
    // Create an rest client not targeting specific API version
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
    }

        factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
        controller := NewNodeLoggingController(factory)
        stop := make(chan struct{})
        defer close(stop)
        err = controller.Run(stop)
        if err != nil {
                log.Fatal(err)
        }
        select {}
}
