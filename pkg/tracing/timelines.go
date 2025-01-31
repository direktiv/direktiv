package tracing

import "errors"

type Timeline struct {
	Meta     MetaData    `json:"meta"`
	Timeline []*SpanNode `json:"timeline"`
}

type MetaData struct {
	TraceID       string `json:"traceID"`
	TotalDuration int64  `json:"totalDurationNano"`
	Status        string `json:"status"`
	Workflow      string `json:"workflow"`
	RootInstance  string `json:"root-instance"`
	Namespace     string `json:"namespace"`
}

type SpanNode struct {
	SpanID    string      `json:"spanId"`
	Name      string      `json:"name"`
	StartTime string      `json:"startTime"`
	EndTime   string      `json:"endTime"`
	Details   Details     `json:"details"`
	Children  []*SpanNode `json:"children,omitempty"`
}

type Details struct {
	Workflow  string `json:"workflow"`
	Instance  string `json:"instance"`
	Namespace string `json:"namespace"`
	Invoker   string `json:"invoker,omitempty"`
	State     string `json:"state,omitempty"`
	Status    string `json:"status,omitempty"`
}

func ConvertTracesToTimelines(traces []map[string]interface{}) (Timeline, error) {
	if len(traces) == 0 {
		return Timeline{}, errors.New("no traces available")
	}

	// Extract metadata from the first trace
	firstTrace := traces[0]
	traceID := safeString(firstTrace, "traceId")
	traceGroupFields := safeMap(firstTrace, "traceGroupFields")
	duration := safeFloat64(traceGroupFields, "durationInNanos")

	timeline := Timeline{
		Meta: MetaData{
			TraceID:       traceID,
			TotalDuration: int64(duration),
			Status:        "completed",
			Workflow:      safeString(firstTrace, "span.attributes.workflow"),
			RootInstance:  safeString(firstTrace, "span.attributes.instance"),
			Namespace:     safeString(firstTrace, "span.attributes.namespace"),
		},
	}

	spanMap := buildSpanMap(traces)

	// Build the tree roots
	rootSpans := []*SpanNode{}
	for _, trace := range traces {
		spanID := safeString(trace, "spanId")
		parentID := safeString(trace, "parentSpanId")

		span, exists := spanMap[spanID]
		if !exists {
			continue
		}

		if parentID == "" {
			// This is a root span
			rootSpans = append(rootSpans, span)
		}
	}

	// Build the leaves
	for _, trace := range traces {
		spanID := safeString(trace, "spanId")
		parentID := safeString(trace, "parentSpanId")

		span, exists := spanMap[spanID]
		if !exists {
			continue
		}

		if parentID != "" {
			// This is a root span
			if parentSpan, parentExists := spanMap[parentID]; parentExists {
				// Attach as a child of the parent span
				parentSpan.Children = append(parentSpan.Children, span)
			}
		}
	}
	timeline.Timeline = rootSpans

	return timeline, nil
}

func buildSpanMap(traces []map[string]interface{}) map[string]*SpanNode {
	spanMap := make(map[string]*SpanNode)

	for _, trace := range traces {
		spanID := safeString(trace, "spanId")
		spanMap[spanID] = &SpanNode{
			SpanID:    spanID,
			Name:      safeString(trace, "name"),
			StartTime: safeString(trace, "startTime"),
			EndTime:   safeString(trace, "endTime"),
			Details: Details{
				Workflow:  safeString(trace, "span.attributes.workflow"),
				Instance:  safeString(trace, "span.attributes.instance"),
				Namespace: safeString(trace, "span.attributes.namespace"),
				Invoker:   safeString(trace, "span.attributes.invoker"),
				State:     safeString(trace, "span.attributes.state"),
				Status:    safeString(trace, "span.attributes.status"),
			},
			Children: []*SpanNode{},
		}
	}

	return spanMap
}

func safeString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}

	return ""
}

func safeMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key].(map[string]interface{}); ok {
		return val
	}

	return map[string]interface{}{}
}

func safeFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}

	return 0
}
