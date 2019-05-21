package controller

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/box-autoremediation/pkg/controller/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Update(client *kubernetes.Clientset, interval float64, ch <-chan types.Alert) {
	buf := make(map[string]string)
	now := time.Now().UTC()
	//configmapClient.
	configmapClient := client.CoreV1().ConfigMaps("node-problem-detector")
	for {
		select {
		// dedupe & update map after timeout
		case r := <-ch:
			buf[r.Node+"_"+string(r.Condition)] = r.Params
			//log.Info("Found issue on ", r.Node, " in update.go")
			if time.Since(now).Seconds() > interval {
				now = time.Now().UTC()
				log.Debug(buf, len(buf))
				//Create config map
				cm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "npd-alerts",
					},
					Data: buf,
				}
				log.Info("Updater - Updating configmap in update.go")
				result, err := configmapClient.Update(cm)
				log.Debug("Updater - Created configmap ", result)
				if err != nil {
					log.Fatal("Updater - Could not update configmap :", err)
				}
				buf = make(map[string]string)
			}
		}
	}
}
