package controller

import (
	"time"
	"strconv"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/box-autoremediation/pkg/controller/types"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" 
)

//CheckCordoned function gets cordon'd nodes periodically,
// filters then one's which have maintenance labels defined and if drainTimestamp is older than 
// defined threhold.
func CheckCordoned(client *kubernetes.Clientset, cond map[string]string, config *viper.Viper, alertCh chan<- types.Alert) {

ticker := time.NewTicker(config.GetDuration("cordoned.check_frequency"))
for {
	select {
	case <-ticker.C:
	nodeList, err  := client.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: "box.com/pool in (generic, calico)", FieldSelector: "spec.unschedulable=true" })
	if err != nil {
		log.Errorf("Cordoned - Could not list nodes: %v",err)
		continue
	}
	for _, node := range nodeList.Items {
		
		if _, ok := node.Labels["maintenance.box.com/source"]; ok {
			cordonTime, err := strconv.Atoi(strings.Split(node.Labels["maintenance.box.com/drainTimestamp"], ".")[0])
			if err != nil {
				log.Errorf("Cordoned - Ignoring %s could not convert maintenance.box.com/drainTimestamp to int: , %v", node.Name, err)
				continue
			}
			timeStamp := time.Unix(int64(cordonTime),0)
			if time.Since(timeStamp) > config.GetDuration("cordoned.duration") {
				log.Infof("Cordoned - Found nodes which cordon'd more than %s ago", config.GetString("cordoned.duration"))
				alertCh <- types.Alert{
								Node: node.Name,
								Condition: "Node-Cordoned",
								Attr: types.Action{
									Timestamp: timeStamp,
									Action: cond["action"],
									Params: "",
									SuccessWait: cond["success_wait"],
									FailedRetry: cond["failed_retry"],
								},
							}
				}
		}
	}
	}
	}
ticker.Stop()
}