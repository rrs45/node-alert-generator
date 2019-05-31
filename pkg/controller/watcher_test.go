package controller

import (
	"reflect"
	"testing"
	"time"

	"github.com/box-autoremediation/pkg/controller/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConditions(t *testing.T) {
	const layout = "2006 - 01 - 02 15: 04: 05 - 0700 MST"
	var nilAlert []types.Alert
	now := time.Now()
	last10mins := now.Add(time.Minute * time.Duration(-10))
	last5mins := now.Add(time.Minute * time.Duration(-5))
	last1min := now.Add(time.Minute * time.Duration(-1))
	last25hour := now.Add(time.Hour * time.Duration(-25))

	frequency, _ := time.ParseDuration("24h")
	cond_table := []struct {
		conds       []v1.NodeCondition
		node        string
		expectedBuf []types.Alert
		isNodeReady bool
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
			expectedBuf: []types.Alert{
				{
					Timestamp: last1min,
					Node:      "fake-compute-node.dsv31.boxdc.net",
					Condition: "NPD-KubeletCertExpiring",
					Action:    "",
					Params:    "status = OK threshold_days = 60 result_days = 280",
				},
			},
			isNodeReady: true,
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
			expectedBuf: []types.Alert{
				{
					Timestamp: last1min,
					Node:      "fake-compute-node.dsv31.boxdc.net",
					Condition: "NPD-KubeletIsDown",
					Action:    "",
					Params:    "status = CRITICAL",
				},
			},
			isNodeReady: false,
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
			expectedBuf: nilAlert,
			isNodeReady: true,
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
			expectedBuf: nilAlert,
			isNodeReady: true,
		},
	}

	for _, item := range cond_table {
		buf, nodeStatus := checkConditions(item.conds, item.node, frequency)
		if !reflect.DeepEqual(buf, item.expectedBuf) {
			t.Errorf("Returned: %+v \n Expected: %+v", buf, item.expectedBuf)
		}
		if item.isNodeReady != nodeStatus {
			t.Error("Node status is not incorrect")
		}
	}

}

func TestLabels(t *testing.T) {
	labels_table := []struct {
		labels map[string]string
		result string
	}{
		{
			labels: map[string]string{
				"maintenance.box.com/source": "npd",
				"box.com/pool":               "calico"},
			result: "npd_maint",
		},
		{
			labels: map[string]string{
				"maintenance.box.com/source": "user_rajsingh",
				"box.com/pool":               "calico"},
			result: "non_npd_maint",
		},
		{
			labels: map[string]string{
				"box.com/calico-pod": "true",
				"box.com/pool":       "calico"},
			result: "no_maint",
		},
	}

	for _, l := range labels_table {
		r := checkLabels(l.labels)
		if r != l.result {
			t.Error("Unexpected result")
		}
	}
}
