package controller

import (
	"regexp"
	"strings"
	"time"
	v1 "k8s.io/api/core/v1"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
	"github.com/box-autoremediation/pkg/controller/types"
)

//Filter filters nodes based on labels, conditions and status
func Filter(filterCh <-chan *v1.Node, labelch chan<- *v1.Node, alertCh chan<- types.Alert, conf *viper.Viper) {
	for {
		select {
		case node := <-filterCh:	
		labelExcludeFilterOK := labelExcludeFilter(node.GetLabels(), conf.Sub("label_filter"))
		labelIncludeFilterOK := labelIncludeFilter(node.GetLabels(), conf.Sub("label_filter"))
		conditions, nodeOK := conditionsFilter(node.Status.Conditions, node.Name, conf.Sub("condition"), conf.GetBool("node_status.include_not_ready"))
		log.Debugf("Filter - %+v", conditions)
		log.Debugf("Filter - %+v", node.GetLabels())
		log.Debugf("Filter - %v , %v",labelExcludeFilterOK, labelIncludeFilterOK)
		if labelExcludeFilterOK && labelIncludeFilterOK && nodeOK && len(conditions) > 0 {
			for _, alert := range conditions {
				log.Infof("Filter - Sending %v_%v to updater", alert.Node, alert.Condition)
				alertCh <- alert
				}
			if conf.GetBool("SetNodeLabel") {
				//log.Infof("Filter - Sending %v to labeller", node.Name)
				labelch <- node
			}
		}
		}
	}
} 

func labelExcludeFilter(labels map[string]string, labelFilter *viper.Viper) (bool) {
	if  !labelFilter.IsSet("exclude.key"){
		//No exclude labels defined, do not ignore node
		return true
	}
	for key, val := range labels {
		if key == labelFilter.GetString("exclude.key") {
			if labelFilter.IsSet("exclude.not_val") && val == labelFilter.GetString("exclude.not_val") {
				//Node has excluded key but included value, do not ignore this node
				return true
			} else if labelFilter.IsSet("exclude.not_match") {
				if matched, _ := regexp.MatchString(labelFilter.GetString("exclude.not_match"), val); matched {
					//Node has excluded key but included regex match value, do not ignore this node
					return true
				}
				
			}
			//Node has excluded key but not included value, ignore this node
			return false
		}
	}
	//Node does not have excluded key, do not ignore
	return true
}

func labelIncludeFilter(labels map[string]string, labelFilter *viper.Viper) (bool) {
	if !labelFilter.IsSet("include"){
		//No include labels defined, do not ignore node
		return true
	}
	for key, val := range labels {
		if key == labelFilter.GetString("include.key") {
			if labelFilter.IsSet("include.match_val") {
				matched, _ := regexp.MatchString(labelFilter.GetString("include.match_val"), val)
				if matched{
					//Node has the include key and matching value, do not ignore this node
					return true
				}
				//Node has the include key but no matching value, ignore this node
				return false
			}
			//Node has include key and no match criteria defined, do not ignore this node
			return true
		}
	}
	//Node does not have the include key, ignore this node
	return false
}

//Get not ready nodes 
func conditionsFilter(conditions []v1.NodeCondition, nodeName string, condFilter *viper.Viper,  inclNotReady bool) ([]types.Alert, bool) {
	var includeNode bool
	var buf []types.Alert
	var item types.Alert

	for _, condition := range conditions {
		condLower := strings.ToLower(string(condition.Type))
		matched, _ := regexp.MatchString(condFilter.GetString("options.match"), string(condition.Type))
		if matched && condition.Status == "True" && time.Since(condition.LastHeartbeatTime.Time) < condFilter.GetDuration("options.interval") && condFilter.IsSet("name." + condLower){
			item.Node = nodeName
			item.Condition = string(condition.Type)
			item.Attr.Timestamp = condition.LastHeartbeatTime.Time
			item.Attr.Params = condition.Message
			item.Attr.Action = condFilter.GetString("name." + string(condition.Type) + ".action" )
			item.Attr.SuccessWait = condFilter.GetString("name." + string(condition.Type) + ".success_wait" )
			item.Attr.FailedRetry = condFilter.GetString("name." + string(condition.Type) + ".failed_retry" )
			buf = append(buf, item)
		} else if condition.Type == "Ready" {
			if  inclNotReady{
				includeNode = true
				continue
			} else if condition.Status == "True"{
				includeNode = true
				continue
			} else {
				includeNode = false
			}
		} 
	}
	if len(buf) == 0 && inclNotReady && condFilter.IsSet("Name.Node-Not-Ready"){
			//Only Not Ready nodes with no failing NPD checks
			item.Node = nodeName
			item.Condition = "Node-Not-Ready"
			item.Attr.Action = condFilter.GetString("Name.Node-Not-Ready.Action" )
			item.Attr.SuccessWait = condFilter.GetString("Name.Node-Not-Ready.SuccessWait" )
			item.Attr.FailedRetry = condFilter.GetString("Name.Node-Not-Ready.FailedRetry" )
			buf = append(buf, item)
	}
return buf, includeNode
}
