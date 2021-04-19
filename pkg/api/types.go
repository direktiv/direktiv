package api

type NameDataTuple struct {
	Name string `json:"name"`
	Data string `json:"data,omitempty"`
}

type RequestStatus struct {
	StatusCode int
	Message    string
}

type JQQuery struct {
	Input map[string]interface{} `json:"input"`
	Query string                 `json:"query"`
}
