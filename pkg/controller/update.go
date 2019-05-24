package controller

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/box-autoremediation/pkg/controller/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Update(client *kubernetes.Clientset, interval string, ch <-chan types.Alert) {
	buf_cur := make(map[string]string)
	buf_prev := make(map[string]string)
	frequency, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatal("Updater - Could not parse interval: ", err)
	}
	ticker := time.NewTicker(frequency)
	configmapClient := client.CoreV1().ConfigMaps("node-problem-detector")

	for {
		select {
		case <-ticker.C:
			log.Debug(buf_cur, len(buf_cur))
			eq := reflect.DeepEqual(buf_prev, buf_cur)
			if eq {
				log.Info("Updater - No new entries found")
			} else {
				//Create config map
				cm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "npd-alerts",
					},
					Data: buf_cur,
				}
				log.Info("Updater - Updating configmap with ", len(buf_cur), " entries")
				for count := 0; count < 3; count++ {
					result, err := configmapClient.Update(cm)
					if err != nil {
						if count < 3 {
							log.Infof("Updater - Could not update configmap tried %d times, retrying after 1000ms: %s", count, err)
							time.Sleep(100 * time.Millisecond)
							continue
						} else {
							log.Errorf("Updater - Could not update configmap after 3 attempts: %s", err)
						}
					} else {
						log.Debug("Updater - Created configmap ", result)
						break
					}
				}

			}
			buf_prev = buf_cur
			buf_cur = make(map[string]string)
		default:
			select {
			case r := <-ch:
				buf_cur[r.Node+"_"+string(r.Condition)] = r.Params
			default:
				continue
			}
		}
	}
	ticker.Stop()
}
