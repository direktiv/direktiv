package tsservice

type Definition struct {
	Type  string
	Store string
	Json  bool
	State string
	Cron  string

	Timeout string

	Event  FlowEvent
	Events []FlowEvent

	Scale []Scale
}

type FlowEvent struct {
	Type    string
	Context map[string]interface{}
}

type Scale struct {
	Min    int
	Max    int
	Cron   string
	Metric string
	Value  int
}

func DefaultDefinition() *Definition {

	return &Definition{
		Type:    defTypeDefault,
		Store:   defStoreAlways,
		Json:    true,
		Timeout: defTimoutDefault,
		Scale: []Scale{
			{
				Min:    0,
				Max:    1,
				Metric: defMetricInstances,
				Value:  100,
			},
		},
	}
}

const (
	defTypeDefault   = "default"
	defTypeScheduled = "scheduled"
	defTypeEvent     = "event"
	defTypeEventsAnd = "eventsAnd"
	defTypeEventsOr  = "eventsOr"

	defTimoutDefault = "PT15M"

	defStoreAlways = "always"
	defStoreError  = "error"
	defStoreNever  = "never"

	defMetricInstances = "instances"
)

func (def *Definition) Validate() *Messages {
	m := newMessages()

	return m
}
