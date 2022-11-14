package main

import "time"

type executeResponse struct {
	Instance string `json:"instance,omitempty"`
}

type logResponse struct {
	PageInfo struct {
		TotalCount int `json:"total"`
		Limit      int `json:"limit"`
		Offset     int `json:"offset"`
	} `json:"pageInfo"`
	Results []struct {
		T   time.Time `json:"t"`
		Msg string    `json:"msg"`
	} `json:"results"`
	Namespace string `json:"namespace"`
	Instance  string `json:"instance"`
}

type instanceResponse struct {
	Namespace string `json:"namespace"`
	Instance  struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		ID           string    `json:"id"`
		As           string    `json:"as"`
		Status       string    `json:"status"`
		ErrorCode    string    `json:"errorCode"`
		ErrorMessage string    `json:"errorMessage"`
	} `json:"instance"`
	InvokedBy string   `json:"invokedBy"`
	Flow      []string `json:"flow"`
	Workflow  struct {
		Path     string `json:"path"`
		Name     string `json:"name"`
		Parent   string `json:"parent"`
		Revision string `json:"revision"`
	} `json:"workflow"`
}

type instanceOutput struct {
	Namespace string `json:"namespace"`
	Instance  struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		ID           string    `json:"id"`
		As           string    `json:"as"`
		Status       string    `json:"status"`
		ErrorCode    string    `json:"errorCode"`
		ErrorMessage string    `json:"errorMessage"`
	} `json:"instance"`
	Data string `json:"data"`
}
