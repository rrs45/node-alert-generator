package controller

import (
	"encoding/json"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/box-autoremediation/pkg/controller/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

//Update creates config map if it doesnt exist and
//updates the config map with alerts received from watcher
func Update(client *kubernetes.Clientset, ns string, configMap string, interval string, ch <-chan types.Alert) {
	bufCur := make(map[string]types.Action)
	buf := make(map[string]string)
	bufPrev := make(map[string]types.Action)
	frequency, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatal("Updater - Could not parse interval: ", err)
	}
	ticker := time.NewTicker(frequency)
	configmapClient := client.CoreV1().ConfigMaps(ns)
	initConfigMap(configmapClient, configMap)
	labelConfigMap(configmapClient, configMap)

	for {
		select {
		case <-ticker.C:
			log.Debugf("%+v %d", bufCur, len(bufCur))
			eq := reflect.DeepEqual(bufPrev, bufCur)
			if eq {
				log.Info("Updater - No new entries found")
			} else {
				//Create config map
				for cond, val := range bufCur {
					rstr, err := json.Marshal(val)
					if err != nil {
						log.Errorf("ConfigMap Updater - unable to marshal %+v: %v", val, err)
					} else {
						buf[cond] = string(rstr)
					}
				}
				cm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: configMap,
					},
					Data: buf,
				}
				log.Info("Updater - Updating configmap with ", len(bufCur), " entries")
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
						log.Debug("Updater - Updated configmap ", result)
						break
					}
				}
				//labelConfigMap(configmapClient, configMap)
			}
			bufPrev = bufCur
			bufCur = make(map[string]types.Action)
		default:
			select {
			case r := <-ch:
				bufCur[r.Node+"_"+string(r.Condition)] = types.Action{
					Timestamp:   r.Attr.Timestamp,
					Action:      r.Attr.Action,
					Params:      r.Attr.Params,
					SuccessWait: r.Attr.SuccessWait,
					FailedRetry: r.Attr.FailedRetry}
			default:
				continue
			}
		}
	}
	ticker.Stop()
}

func initConfigMap(configmapClient corev1.ConfigMapInterface, name string) {
	_, err1 := configmapClient.Get(name, metav1.GetOptions{})
	if err1 != nil {
		log.Infof("Updater - %s configmap not found, creating new one", name)
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		for count := 0; count < 3; count++ {
			result, err2 := configmapClient.Create(cm)
			if err2 != nil {
				if count < 3 {
					log.Infof("Updater - Could not create configmap tried %d times, retrying after 1000ms: %s", count, err2)
					time.Sleep(1000 * time.Millisecond)
					continue
				} else {
					log.Errorf("Updater - Could not create configmap after 3 attempts: %s", err2)
				}
			} else {
				log.Debug("Updater - Created configmap ", result)
				break
			}
		}
	} else {
		log.Infof("Updater - %s configmap already exists", name)
	}
}

func labelConfigMap(configmapClient corev1.ConfigMapInterface, name string) {
	foundCM, err := configmapClient.Get(name, metav1.GetOptions{})
	if err != nil {
		log.Infof("Updater - %s configmap not found", name)
	} else {
		oldData, err1 := json.Marshal(foundCM)
		if err1 != nil {
			log.Error("Labeler - could not marshal old node object", err1)
		}
		foundCM.SetLabels(map[string]string{"autoremediation": "yes"})
		newData, err := json.Marshal(foundCM)
		if err != nil {
			log.Error(err, "Updater - could not marshal new node object")
		}
		patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, foundCM)
		if err != nil {
			log.Error("Updater - could not create two way merge patch ", err)
		}
		log.Info("Updater - Update label for ", foundCM.Name)
		_, err = configmapClient.Patch(foundCM.Name, k8types.MergePatchType, patch)
		if err != nil {
			log.Error("Updater -  could not patch configmap", err)
		}
	}
}
