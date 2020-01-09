package controller

import (
	"reflect"
	"testing"
	"time"

	"github.com/box-autoremediation/pkg/controller/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/spf13/viper"
)

func TestConditions(t *testing.T) {
	const layout = "2006 - 01 - 02 15: 04: 05 - 0700 MST"
	var nilAlert []types.Alert
	now := time.Now()
	last10mins := now.Add(time.Minute * time.Duration(-10))
	last5mins := now.Add(time.Minute * time.Duration(-5))
	last1min := now.Add(time.Minute * time.Duration(-1))
	last25hour := now.Add(time.Hour * time.Duration(-25))

	condOptions := viper.New()
	condOptions.Set("options.match", "NPD-.*")
	condOptions.Set("options.interval","24h")
	condOptions.Set("name.npd-kubeletcertexpiring.action", "test1.yml")
	condOptions.Set("name.npd-kubeletcertexpiring.success_wait", "1h")
	condOptions.Set("name.npd-kubeletcertexpiring.failed_retry", "3")
	condOptions.Set("name.npd-kubeletisdown.action", "test1.yml")
	condOptions.Set("name.npd-kubeletisdown.success_wait", "1h")
	condOptions.Set("name.npd-kubeletisdown.failed_retry", "3")
	//condOptions.Set("condition.Name.NPD-KubeletCertExpiring", map[string]string{"action": "test1.yml"})
	t.Logf("%+v",condOptions)

	//frequency, _ := time.ParseDuration("24h")
	condTable := []struct {
		conds       []v1.NodeCondition
		node        string
		inclNotReady bool
		expectedBuf []types.Alert
		expectedOK bool
		
	}{
		{
			conds: []v1.NodeCondition{
				{
					Type:               "Ready",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last10mins},
					Reason:             "KubeletIsHealthy",
					Message:            "Kubelet is healthy",
				},
				{
					Type:               "NPD-KubeletCertExpiring",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last10mins},
					Reason:             "CertExpiring",
					Message:            "status = OK threshold_days = 60 result_days = 280",
				},
			},
			node: "fake-compute-node.dsv31.boxdc.net",
			inclNotReady: false,
			expectedBuf: []types.Alert{
				{
					Node:      "fake-compute-node.dsv31.boxdc.net",
					Condition: "NPD-KubeletCertExpiring",
					Attr:     types.Action{
									Action: "test1.yml",
									Timestamp: last1min,
									Params:    "status = OK threshold_days = 60 result_days = 280",
									SuccessWait: "1h",
									FailedRetry: "3",},
				},
			},
			expectedOK: true,	
		},
		{
			conds: []v1.NodeCondition{
				{
					Type:               "Ready",
					Status:             v1.ConditionStatus("False"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsHealthy",
					Message:            "Kubelet is healthy",
				},
				{
					Type:               "NPD-KubeletIsDown",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsDown",
					Message:            "status = CRITICAL",
				},
			},
			node: "fake-compute-node.dsv31.boxdc.net",
			inclNotReady: true,
			expectedBuf: []types.Alert{
				{
					Node:      "fake-compute-node.dsv31.boxdc.net",
					Condition: "NPD-KubeletIsDown",
					Attr:	types.Action{
								Action: "test1.yml",
								Timestamp: last1min,	
								Params:    "status = CRITICAL",
								SuccessWait: "1h",
								FailedRetry: "3",},
				},
			},
			expectedOK: true,
		},
		{
			conds: []v1.NodeCondition{
				{
					Type:               "Ready",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsHealthy",
					Message:            "Kubelet is healthy",
				},
				{
					Type:               "NPD-KubeletIsDown",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last25hour},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsDown",
					Message:            "status = CRITICAL",
				},
			},
			node:        "fake-compute-node.dsv31.boxdc.net",
			inclNotReady: true,
			expectedBuf: nilAlert,		
			expectedOK: true,
		},
		{
			conds: []v1.NodeCondition{
				{
					Type:               "Ready",
					Status:             v1.ConditionStatus("True"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsHealthy",
					Message:            "Kubelet is healthy",
				},
				{
					Type:               "NPD-KubeletIsDown",
					Status:             v1.ConditionStatus("False"),
					LastHeartbeatTime:  metav1.Time{last1min},
					LastTransitionTime: metav1.Time{last5mins},
					Reason:             "KubeletIsDown",
					Message:            "status = OK",
				},
			},
			node:        "fake-compute-node.dsv31.boxdc.net",
			inclNotReady: true,
			expectedBuf: nilAlert,
			expectedOK: true,
		},
	}

	for _, item := range condTable {
		buf, nodeStatus := conditionsFilter(item.conds, item.node, condOptions, item.inclNotReady)
		//t.Logf("%+v, %v",buf, item.expectedBuf)
		if !reflect.DeepEqual(buf, item.expectedBuf) {
			t.Errorf("Returned: %+v \n Expected: %+v", buf, item.expectedBuf)
		}
		if item.expectedOK != nodeStatus {
			t.Error("Node status is not incorrect")
		}
	}

}

func TestLabelExcludeFilter(t *testing.T) {
	exclLabelOptions := viper.New()
	exclLabelOptions.SetDefault("exclude.key", "maintenance.box.com/source")
	exclLabelOptions.SetDefault("exclude.not_val", "npd")
	exclLabelOptions.SetDefault("exclude.not_match", "node-alert-worker.*")


	labelsTable := []struct {
		labels map[string]string
		result bool
	}{
		{
			labels: map[string]string{
				"maintenance.box.com/source": "npd",
				"box.com/pool":               "calico"},
			result: true,
		},
		{
			labels: map[string]string{
				"maintenance.box.com/source": "user_rajsingh",
				"box.com/pool":               "calico"},
			result: false,
		},
		{
			labels: map[string]string{
				"box.com/calico-pod": "true",
				"box.com/pool":       "calico"},
			result: true,
		},
		{
			labels: map[string]string{
				"maintenance.box.com/source": "node-alert-worker-2551377941-g2x3j",
				"box.com/pool":       "calico"},
			result: true,
		},
		{
			labels: map[string]string{
				"maintenance.box.com/source": "node-worker-2551377941-g2x3j",
				"box.com/pool":       "calico"},
			result: false,
		},
	}

	for _, l := range labelsTable {
		r := labelExcludeFilter(l.labels, exclLabelOptions)
		//t.Logf("%v", r)
		if r != l.result {
			t.Error("Unexpected result")
		}
	}
}

func TestLabelIncludeFilter(t *testing.T) {
	inclLabelOptions := viper.New()
	inclLabelOptions.SetDefault("include.key", "box.com/pool")
	inclLabelOptions.SetDefault("include.match_val", "calico")

	t.Logf("%+v",inclLabelOptions)

	labelsTable := []struct {
		labels map[string]string
		result bool
	}{
		{
			labels: map[string]string{
				"maintenance.box.com/source": "node-alert-worker-2289751211-qhmkd",
				"box.com/pool":               "master"},
			result: false,
		},
		{
			labels: map[string]string{
				"maintenance.box.com/source": "user_rajsingh",
				"box.com/pool":               "calico"},
			result: true,
		},
		{
			labels: map[string]string{
				"box.com/calico-pod": "true",
				"box.com/pool":       "calico"},
			result: true,
		},
	}

	for _, l := range labelsTable {
		r := labelIncludeFilter(l.labels, inclLabelOptions)
		//t.Logf("%v", r)
		if r != l.result {
			t.Error("Unexpected result")
		}
	}
}
