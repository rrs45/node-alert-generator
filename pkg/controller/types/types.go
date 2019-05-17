package types

import (
	"k8s.io/api/core/v1"
	"time"
)

//Defines format for alerts detected from NPD
type Alert struct {
	Empty     bool
	Timestamp time.Time
	Node      string
	Condition v1.NodeConditionType
	Action    string
	Params    string
}
