package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// createDefaultNameDir - Creates a NamedDirectory object for default, and makes the directory on fs
func createDefaultNameDir(dir string) (NamedDirectory, error) {
	p := filepath.Join(os.TempDir(), dir)
	defaultDir := NamedDirectory{
		Label:     "default",
		Directory: p,
	}

	err := os.MkdirAll(p, 0744)
	return defaultDir, err
}

// readDirOfType - Reads directory and returns list of file names of a specific file type suffix
func readDirOfType(dirPath, fileTypeSuffix string) ([]string, error) {
	fis, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), fileTypeSuffix) {
			out = append(out, strings.TrimSuffix(fi.Name(), fileTypeSuffix))
		}
	}

	return out, nil
}

func writeJSONResponse(w http.ResponseWriter, obj interface{}) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = io.Copy(w, bytes.NewReader(b)); err != nil {
		return err
	}

	return nil
}

func (s *Server) initTemplateFolders() error {

	if s.cfg.Templates.WorkflowTemplateDirectories == nil {
		s.cfg.Templates.WorkflowTemplateDirectories = make([]NamedDirectory, 0)
	}
	if s.cfg.Templates.ActionTemplateDirectories == nil {
		s.cfg.Templates.ActionTemplateDirectories = make([]NamedDirectory, 0)
	}

	if !s.cfg.hasWorkflowTemplateDefault() {
		defaultDir, err := createDefaultNameDir("workflow-templates")
		if err != nil {
			return err
		}

		s.cfg.Templates.WorkflowTemplateDirectories = append(s.cfg.Templates.WorkflowTemplateDirectories, defaultDir)
	}
	if !s.cfg.hasActionTemplateDefault() {
		defaultDir, err := createDefaultNameDir("action-templates")
		if err != nil {
			return err
		}

		s.cfg.Templates.ActionTemplateDirectories = append(s.cfg.Templates.ActionTemplateDirectories, defaultDir)
	}

	s.wfTemplateDirsPaths = make(map[string]string)
	s.wfTemplateDirs = make([]string, 0)

	for _, fi := range s.cfg.Templates.WorkflowTemplateDirectories {
		s.wfTemplateDirs = append(s.wfTemplateDirs, fi.Label)
		s.wfTemplateDirsPaths[fi.Label] = fi.Directory
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
func (s *Server) workflowTemplateFolders() []string {
	return s.wfTemplateDirs
}

func (s *Server) workflowTemplates(folder string) ([]string, error) {

	// confirm that the specified folder is a 'template directory'
	path, ok := s.wfTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown workflow folder: '%s'", folder)
	}

	return readDirOfType(path, ".yml")
}

func (s *Server) workflowTemplate(folder, name string) ([]byte, error) {

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

func (s *Server) actionTemplateFolders() []string {
	return s.actionTemplateDirs
}

func (s *Server) actionTemplates(folder string) ([]string, error) {

	// confirm that the specified folder is a 'template directory'
	path, ok := s.actionTemplateDirsPaths[folder]
	if !ok {
		return nil, fmt.Errorf("unknown actions folder: '%s'", folder)
	}

	return readDirOfType(path, ".json")
}

func (s *Server) actionTemplate(folder, name string) ([]byte, error) {

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

func (h *Handler) templateFolders(w http.ResponseWriter, r *http.Request) {

	var x interface{}

	switch mux.CurrentRoute(r).GetName() {
	case RN_ListWorkflowTemplateFolders:

		x = h.s.workflowTemplateFolders

	case RN_ListActionTemplateFolders:

		x = h.s.actionTemplateFolders

	default:

		ErrResponse(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)))
		return

	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(x)
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) workflowTemplates(w http.ResponseWriter, r *http.Request) {

	folder := mux.Vars(r)["folder"]

	out, err := h.s.workflowTemplates(folder)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	if err := writeJSONResponse(w, out); err != nil {
		ErrResponse(w, err)
		return
	}
}

func (h *Handler) getTemplate(w http.ResponseWriter, r *http.Request) {

	folder := mux.Vars(r)["folder"]
	n := mux.Vars(r)["template"]

	var err error
	var b []byte

	switch mux.CurrentRoute(r).GetName() {
	case RN_GetWorkflowTemplate:

		b, err = h.s.workflowTemplate(folder, n)
		if err != nil {
			ErrResponse(w, err)
			return
		}

	case RN_GetActionTemplate:

		b, err = h.s.actionTemplate(folder, n)
		if err != nil {
			ErrResponse(w, err)
			return
		}

	default:

		ErrResponse(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)))
		return

	}

	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		ErrResponse(w, err)
		return
	}

}

// --

func (h *Handler) actionTemplates(w http.ResponseWriter, r *http.Request) {

	folder := mux.Vars(r)["folder"]

	out, err := h.s.actionTemplates(folder)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	if err := writeJSONResponse(w, out); err != nil {
		ErrResponse(w, err)
		return
	}
}
