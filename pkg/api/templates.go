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

const DirektivActionsURL = "https://api.github.com/repos/vorteil/direktiv-apps/contents/.direktiv/"
const DirektivWorkflowTemplatesURL = "https://api.github.com/repos/vorteil/direktiv-apps/contents/.direktiv/.templates"

func (s *Server) initGitTemplates() error {

	s.workflowTemplates = make([]string, 0)
	s.workflowTemplateData = make(map[string][]byte)
	s.workflowTemplateInfo = make(map[string]GithubFileInfo)

	err := s.getWorkflowTemplates()
	if err != nil {
		return err
	}

	s.actionTemplates = make([]string, 0)
	s.actionTemplateData = make(map[string][]byte)
	s.actionTemplateInfo = make(map[string]GithubFileInfo)

	err = s.getActionTemplates()
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

	go func() {
		for {
			time.Sleep(time.Minute * 60)
			err := s.getActionTemplates()
			if err != nil {
				// TODO log error
			}
		}
	}()

	return nil
}

func (s *Server) getActionTemplates() error {

	u := fmt.Sprintf(DirektivActionsURL)
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
		if fi.Type == "file" {
			n := strings.TrimSuffix(fi.Name, ".json")
			list = append(list, n)
			dataMap[n] = fi
		}
	}

	sort.Strings(list)

	s.actionTemplatesMutex.Lock()
	defer s.actionTemplatesMutex.Unlock()

	s.actionTemplates = list
	s.actionTemplateInfo = dataMap
	s.actionTemplateData = make(map[string][]byte)

	return nil
}

func (s *Server) getActionTemplate(fi GithubFileInfo) ([]byte, error) {

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

	s.actionTemplateData[strings.TrimSuffix(fi.Name, ".json")] = data
	return data, nil
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

func (s *Server) ActionTemplates() []string {
	return s.actionTemplates
}

func (s *Server) WorkflowTemplates() []string {
	return s.workflowTemplates
}

func (s *Server) ActionTemplate(name string) ([]byte, error) {

	s.actionTemplatesMutex.Lock()
	defer s.actionTemplatesMutex.Unlock()

	fi, ok := s.actionTemplateInfo[name]
	if !ok {
		return nil, fmt.Errorf("unknown template '%s'", name)
	}

	if b, ok := s.actionTemplateData[name]; ok {
		return b, nil
	}

	// attempt to get data from GitHub, otherwise fail
	b, err := s.getActionTemplate(fi)
	if err != nil {
		return nil, err
	}

	return b, nil
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
