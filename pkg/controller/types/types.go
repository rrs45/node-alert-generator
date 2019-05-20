package types

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

//Defines format for alerts detected from NPD
type Alert struct {
	Timestamp time.Time
	Node      string
	Condition v1.NodeConditionType
	Action    string
	Params    string
}
