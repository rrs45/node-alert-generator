package controller

import (
	"encoding/json"
	"sync"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

func LabelNode(client *kubernetes.Clientset, wg *sync.WaitGroup, ch <-chan *v1.Node) {

LOOP:
	for {
		select {
		case n := <-ch:
			log.Info("Received item on label channel")
			oldData, err := json.Marshal(n)
			if err != nil {
				log.Error(err, "could not marshal old node object")
			}
			l := n.GetLabels()
			if _, ok := l["maintenance.box.com/source"]; ok {
				log.Info("Label exists hence ignoring in labeller.go")
				goto LOOP
			}
			l["maintenance.box.com/source"] = "npd"
			n.SetLabels(l)
			newData, err := json.Marshal(n)
			if err != nil {
				log.Error(err, "could not marshal new node object in labeller.go")
			}

			patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, n)
			if err != nil {
				log.Error(err, "could not create two way merge patch in labeller.go")
			}
			log.Info("Update label for ", n.Name)
			_, err = client.CoreV1().Nodes().Patch(n.Name, types.MergePatchType, patch)
			if err != nil {
				log.Error(err, "could not patch node in labeller.go")
			}
		}
	}
	wg.Done()

}
