package controller

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

//LabelNode labels the node with given maintenace labels
func LabelNode(client *kubernetes.Clientset, ch <-chan *v1.Node) {
	for {
		select {
		case n := <-ch:
			log.Info("Labeller - Received item on label channel")
			oldData, err := json.Marshal(n)
			if err != nil {
				log.Error("Labeler - could not marshal old node object", err)
			}
			l := n.GetLabels()
			if _, ok := l["maintenance.box.com/source"]; ok {
				log.Info("Labeller - Label exists hence ignoring")
				continue
			}
			l["maintenance.box.com/source"] = "npd"
			n.SetLabels(l)
			newData, err := json.Marshal(n)
			if err != nil {
				log.Error(err, "Labeler - could not marshal new node object")
			}

			patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, n)
			if err != nil {
				log.Error("Labeler - could not create two way merge patch ", err)
			}
			log.Info("Update label for ", n.Name)
			_, err = client.CoreV1().Nodes().Patch(n.Name, types.MergePatchType, patch)
			if err != nil {
				log.Error("Labeler - could not patch node ", err)
			}
		}
	}

}
