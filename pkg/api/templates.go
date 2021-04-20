package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) initWorkflowTemplates() error {
	if s.cfg.Templates.WorkflowTemplateDirectories == nil {
		s.cfg.Templates.WorkflowTemplateDirectories = make([]NamedDirectory, 0)
	}

	var hasDefault bool
	for _, tuple := range s.cfg.Templates.WorkflowTemplateDirectories {
		if tuple.Label == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		p := filepath.Join(os.TempDir(), "workflow-templates")
		s.cfg.Templates.WorkflowTemplateDirectories = append(s.cfg.Templates.WorkflowTemplateDirectories, NamedDirectory{
			Label:     "default",
			Directory: p,
		})

		err := os.MkdirAll(p, 0744)
		if err != nil {
			return err
		}
	}

	s.wfTemplateDirsPaths = make(map[string]string)
	s.wfTemplateDirs = make([]string, 0)

	for _, fi := range s.cfg.Templates.WorkflowTemplateDirectories {
		s.wfTemplateDirs = append(s.wfTemplateDirs, fi.Label)
		s.wfTemplateDirsPaths[fi.Label] = fi.Directory
	}

	return nil
}

func (s *Server) initActionTemplates() error {
	if s.cfg.Templates.ActionTemplateDirectories == nil {
		s.cfg.Templates.ActionTemplateDirectories = make([]NamedDirectory, 0)
	}

	var hasDefault bool
	for _, tuple := range s.cfg.Templates.ActionTemplateDirectories {
		if tuple.Label == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		p := filepath.Join(os.TempDir(), "action-templates")
		s.cfg.Templates.ActionTemplateDirectories = append(s.cfg.Templates.ActionTemplateDirectories, NamedDirectory{
			Label:     "default",
			Directory: p,
		})

		err := os.MkdirAll(p, 0744)
		if err != nil {
			return err
		}
	}

	s.actionTemplateDirsPaths = make(map[string]string)
	s.actionTemplateDirs = make([]string, 0)

	for _, fi := range s.cfg.Templates.ActionTemplateDirectories {
		s.actionTemplateDirs = append(s.actionTemplateDirs, fi.Label)
		s.actionTemplateDirsPaths[fi.Label] = fi.Directory
	}

	return nil
}

// WorkflowTemplates returns a list of files found within the
// directory specified in the Server Config, returning only
// files that bare the suffix ".yml".
func (s *Server) WorkflowTemplateFolders() []string {
	return s.wfTemplateDirs
}

func (s *Server) WorkflowTemplates(folder string) ([]string, error) {

	// confirm that the specified folder is a 'template directory'
	path, ok := s.wfTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown workflow folder: '%s'", folder)
	}

	fis, err := ioutil.ReadDir(path)
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

func (s *Server) WorkflowTemplate(folder, name string) ([]byte, error) {

	path, ok := s.wfTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown workflow folder: '%s'", folder)
	}

	b, err := ioutil.ReadFile(filepath.Join(path, fmt.Sprintf("%s.yml", name)))
	if err != nil {
		return nil, err
	}

	return b, nil
}

// --

func (s *Server) ActionTemplateFolders() []string {
	return s.actionTemplateDirs
}

func (s *Server) ActionTemplates(folder string) ([]string, error) {

	// confirm that the specified folder is a 'template directory'
	path, ok := s.actionTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown actions folder: '%s'", folder)
	}

	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".json") {
			out = append(out, strings.TrimSuffix(fi.Name(), ".json"))
		}
	}

	return out, nil
}

func (s *Server) ActionTemplate(folder, name string) ([]byte, error) {

	path, ok := s.actionTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown actions folder: '%s'", folder)
	}

	b, err := ioutil.ReadFile(filepath.Join(path, fmt.Sprintf("%s.json", name)))
	if err != nil {
		return nil, err
	}

	return b, nil
}
