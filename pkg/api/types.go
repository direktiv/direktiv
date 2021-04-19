package api

type NameDataTuple struct {
	Name string `json:"name"`
	Data string `json:"data,omitempty"`
}

type RequestStatus struct {
	StatusCode int
	Message    string
}
