package types

import (
	"time"
)

//Action defines format for alerts detected from NPD
type Action struct {
	Timestamp time.Time
	Action    string
	Params    string
	SuccessWait string
	FailedRetry string
}

//Alert defines format for alerts detected from NPD
type Alert struct {
	Node      string
	Condition string
	Attr	Action
}
