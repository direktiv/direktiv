package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/project"
	"github.com/gobwas/glob"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	DefaultProfileConfigName = ".direktiv.profile.yaml"
	DefaultProfileConfigPath = ".config/direktiv/"
)

type ProfileConfig struct {
	ID        string `yaml:"id" mapstructure:"profile"`
	Addr      string `yaml:"addr" mapstructure:"addr"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`
	Auth      string `yaml:"auth" mapstructure:"auth"`
	MaxSize   int64  `yaml:"max-size" mapstructure:"max-size"`
}

type Configuration struct {
	ProfileConfig    `yaml:",inline" mapstructure:",squash"`
	project.Config   `yaml:",inline" mapstructure:",squash"`
	Profiles         map[string]ProfileConfig `yaml:"profiles,flow" mapstructure:"profiles"`
	currentProfileID string
	projectPath      string
}

var (
	config   Configuration
	Globbers []glob.Glob
)

func initCLI(cmd *cobra.Command) {
	chdir, err := cmd.Flags().GetString("directory")
	if err != nil {
		Fail("error loading 'directory' flag: %v", err)
	}

	if chdir != "" && chdir != "." {
		err = os.Chdir(chdir)
		if err != nil {
			Fail("error changing directory: %v", err)
		}
	}

	projectPath, err := findProjectDir()
	if err != nil {
		Fail("Failed to read direktiv project config: %v", err)
	}
	err = loadProjectConfig(projectPath)
	if err != nil {
		Fail("Failed to read direktiv project config: %v", err)
	}
	Globbers = make([]glob.Glob, 0)
	for idx, pattern := range config.Ignore {
		g, err := glob.Compile(pattern)
		if err != nil {
			Fail("Failed to parse %dth ignore pattern: %w", idx, err)
		}
		Globbers = append(Globbers, g)
	}

	profileID, err := cmd.Flags().GetString("profile")
	if err != nil {
		Fail("error loading 'profile' flag: %v", err)
	}

	config.currentProfileID = profileID

	err = loadProfileConfig()
	if err != nil {
		Fail("Failed to read profile config file: %v", err)
	}

	if config.currentProfileID == "" {
		for k := range config.Profiles {
			config.currentProfileID = k
			break
		}
	}
	cp, ok := (config.Profiles[config.currentProfileID])
	if !ok {
		Fail("error loading profile '%s': no profile exists by this name in the config file", config.currentProfileID)
	}

	config.projectPath = projectPath

	data, err := yaml.Marshal(cp)
	if err != nil {
		panic(err)
	}

	viper.SetConfigType("yml")

	err = viper.ReadConfig(bytes.NewReader(data))
	if err != nil {
		Fail("error reading config: %v", err)
	}
}

func findProjectDir() (string, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		Fail("Failed to locate place in filesystem: %v\n", err)
	}

	for prev := ""; dir != prev; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, project.ConfigFileName)

		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		prev = dir
	}

	return "", fmt.Errorf("this or any parent folder is not part of a direktiv project")
}

func loadProjectConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config.Config)
	if err != nil {
		return err
	}
	return nil
}

func loadProfileConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		Fail("Could not find user home: %v", err)
	}
	path := filepath.Join(home, DefaultProfileConfigPath+DefaultProfileConfigName)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config.Profiles)
	if err != nil {
		return err
	}
	return nil
}

func getAddr() string {
	addr := viper.GetString("addr")
	if addr == "" {
		Fail("addr undefined: ensure it is set as a flag, environment variable, or in the config file")
	}

	return addr
}

func GetNamespace() string {
	namespace := viper.GetString("namespace")
	if namespace == "" {
		Fail("namespace undefined: ensure it is set as a flag, environment variable, or in the config file")
	}

	return namespace
}

func getInsecure() bool {
	return viper.GetBool("insecure")
}

func GetTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: getInsecure()} //nolint:gosec
}

func GetAuth() string {
	return viper.GetString("auth")
}

func AddAuthHeaders(req *http.Request) {
	req.Header.Add("Direktiv-Token", GetAuth())
}

func AddSSEAuthHeaders(client *sse.Client) {
	client.Headers["Direktiv-Token"] = GetAuth()
}

func GetRelativePath(configPath, targpath string) string {
	var err error

	if !filepath.IsAbs(configPath) {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			Fail("Failed to determine absolute path: %v", err)
		}
	}

	if !filepath.IsAbs(targpath) {
		targpath, err = filepath.Abs(targpath)
		if err != nil {
			Fail("Failed to determine absolute path: %v", err)
		}
	}

	s, err := filepath.Rel(configPath, targpath)
	if err != nil {
		Fail("Failed to generate relative path: %v", err)
	}

	path := filepath.ToSlash(s)
	path = strings.Trim(path, "/")

	return path
}

func GetConfigPath() string {
	if config.projectPath != "" {
		projectPath := config.projectPath
		projectPath = filepath.Dir(projectPath)
		projectPath = strings.TrimSuffix(projectPath, "/")
		return projectPath
	}

	return "."
}
