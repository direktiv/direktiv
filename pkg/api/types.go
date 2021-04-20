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

// GithubDirectoryInfo ..
type GithubFileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

type NamedDirectory struct {
	Label     string
	Directory string
}
