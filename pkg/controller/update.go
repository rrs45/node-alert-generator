package controller

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/box-autoremediation/pkg/controller/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Update(client *kubernetes.Clientset, ch <-chan types.Alert) error {
	flog := log.WithFields(log.Fields{
		"file": "pkg/controller/update.go",
	})
	fmt.Println("update.go")
	//buf := make(map[string]types.Alert)
	buf := make(map[string]string)
	now := time.Now().UTC()
	for {
		select {
		// dedupe & update map after timeout
		case r := <-ch:
			buf[r.Node+"_"+string(r.Condition)] = r.Params
			if time.Since(now).Seconds() > 10 {
				now = time.Now().UTC()
				fmt.Println(buf, len(buf))
				fmt.Println()
				//https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
				configmapClient := client.CoreV1().ConfigMaps("node-problem-detector")
				//configmapClient.
				cm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "npd-alerts",
					},
					Data: buf,
				}
				_, err := configmapClient.Update(cm)
				if err != nil {
					flog.WithFields(log.Fields{
						"function": "configmapClient.Update(cm)",
					}).Fatal(err)
				}
				//fmt.Println(result)
				buf = make(map[string]string)

			}
		}
	}
}
