package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const DirektivAppsURL = "https://api.github.com/repos/vorteil/direktiv-apps/contents/.direktiv/.apps"
const DirektivWorkflowTemplatesURL = "https://api.github.com/repos/vorteil/direktiv-apps/contents/.direktiv/.templates"

func (s *Server) initGitTemplates() error {

	s.workflowTemplates = make([]string, 0)
	s.workflowTemplateData = make(map[string][]byte)
	s.workflowTemplateInfo = make(map[string]GithubFileInfo)

	err := s.getWorkflowTemplates()
	if err != nil {
		return err
	}

	go func() {
		for {
			time.Sleep(time.Minute * 60)
			err := s.getWorkflowTemplates()
			if err != nil {
				// TODO log error
			}
		}
	}()

	return nil
}

func (s *Server) getWorkflowTemplates() error {

	u := fmt.Sprintf(DirektivWorkflowTemplatesURL)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	if s.cfg.Github.UseAuthentication {
		req.SetBasicAuth(s.cfg.Github.Username, s.cfg.Github.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fis := make([]GithubFileInfo, 0)
	err = json.Unmarshal(data, &fis)
	if err != nil {
		return err
	}

	list := make([]string, 0)
	dataMap := make(map[string]GithubFileInfo)

	for _, fi := range fis {
		n := strings.TrimSuffix(fi.Name, ".yml")
		list = append(list, n)
		dataMap[n] = fi
	}

	sort.Strings(list)

	s.workflowTemplatesMutex.Lock()
	defer s.workflowTemplatesMutex.Unlock()

	s.workflowTemplates = list
	s.workflowTemplateInfo = dataMap
	s.workflowTemplateData = make(map[string][]byte)

	return nil
}

func (s *Server) getWorkflowTemplate(fi GithubFileInfo) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, fi.DownloadURL, nil)
	if err != nil {
		return nil, err
	}
	if s.cfg.Github.UseAuthentication {
		req.SetBasicAuth(s.cfg.Github.Username, s.cfg.Github.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github responded with a non-200 status code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s.workflowTemplateData[strings.TrimSuffix(fi.Name, ".yml")] = data
	return data, nil
}

// WorkflowTemplates returns a list of files found within the
// directory specified in the Server Config, returning only
// files that bare the suffix ".yml".
func (s *Server) WorkflowTemplates() []string {
	return s.workflowTemplates
}

func (s *Server) WorkflowTemplate(name string) ([]byte, error) {

	s.workflowTemplatesMutex.Lock()
	defer s.workflowTemplatesMutex.Unlock()

	fi, ok := s.workflowTemplateInfo[name]
	if !ok {
		return nil, fmt.Errorf("unknown template '%s'", name)
	}

	if b, ok := s.workflowTemplateData[name]; ok {
		return b, nil
	}

	// attempt to get data from GitHub, otherwise fail
	b, err := s.getWorkflowTemplate(fi)
	if err != nil {
		return nil, err
	}

	return b, nil
}
