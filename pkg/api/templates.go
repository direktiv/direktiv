package api

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// WorkflowTemplates returns a list of files found within the
// directory specified in the Server Config, returning only
// files that bare the suffix ".yml".
func (s *Server) WorkflowTemplates() ([]string, error) {

	fis, err := ioutil.ReadDir(s.cfg.Templates.WorkflowsDirectory)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".yml") {
			out = append(out, strings.TrimSuffix(fi.Name(), ".yml"))
		}
	}

	return out, nil
}

func (s *Server) WorkflowTemplate(name string) ([]byte, error) {

	b, err := ioutil.ReadFile(filepath.Join(s.cfg.Templates.WorkflowsDirectory, fmt.Sprintf("%s.yml", name)))
	if err != nil {
		return nil, err
	}

	return b, nil
}
