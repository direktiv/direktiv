package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	DefaultProfileConfigName = ".direktiv.profile.yaml"
	DefaultProfileConfigPath = ".config/direktiv/"
	ConfigFileName           = ".direktiv.yaml"
)

type ProfileConfig struct {
	Addr      string `mapstructure:"addr"      yaml:"addr"`
	Namespace string `mapstructure:"namespace" yaml:"namespace"`
	Auth      string `mapstructure:"auth"      yaml:"auth"`
	MaxSize   int64  `mapstructure:"max-size"  yaml:"max-size"`
}

type ProjectConfig struct {
	Ignore []string `yaml:"ignore"`
}

type Configuration struct {
	ProjectConfig `mapstructure:"config,squash" yaml:"config,inline"`
	Profiles      map[string]ProfileConfig `mapstructure:"profiles"      yaml:"profiles,flow"`
}

var (
	Config   Configuration
	Globbers []glob.Glob
)

func initCLI(cmd *cobra.Command) error {
	cp, err := getCurrentProfileConfig(cmd)
	if err != nil {
		return fmt.Errorf("error initializing %w", err)
	}

	data, err := yaml.Marshal(cp)
	if err != nil {
		panic(err)
	}

	viper.SetConfigType("yml")

	err = viper.ReadConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error reading configuration: %w", err)
	}
	return nil
}

func initProjectDir(cmd *cobra.Command) error {
	chdir, err := cmd.Flags().GetString("directory")
	if err != nil {
		return fmt.Errorf("error loading 'directory' flag: %w", err)
	}

	if chdir != "" && chdir != "." {
		err = os.Chdir(chdir)
		if err != nil {
			return fmt.Errorf("error changing to directory %s: %w", chdir, err)
		}
	}

	projectFile, err := findProjectFile()
	if err != nil {
		return fmt.Errorf("unable to find project folder: %w", err)
	}
	err = loadProjectConfig(projectFile)
	if err != nil {
		return fmt.Errorf("failed to read direktiv project configuration-file: %w", err)
	}
	viper.Set("projectFile", projectFile)
	Globbers = make([]glob.Glob, 0)
	for idx, pattern := range Config.Ignore {
		g, err := glob.Compile(pattern)
		if err != nil {
			return fmt.Errorf("failed to parse %dth entry of the ignore pattern: %w", idx, err)
		}
		Globbers = append(Globbers, g)
	}
	return nil
}

func getCurrentProfileConfig(cmd *cobra.Command) (ProfileConfig, error) {
	profileID := viper.GetString("profile")

	err := LoadProfileConfig()
	if err != nil && (getAddr() == "" || GetNamespace() == "") {
		return ProfileConfig{}, fmt.Errorf("failed to read profile config file: %w. Create a profile-config file or specify the addr and namespace via flags", err)
	}
	if err != nil && getAddr() != "" && GetNamespace() != "" {
		return ProfileConfig{
			Addr:      getAddr(),
			Namespace: GetNamespace(),
			Auth:      GetAuth(),
		}, nil
	}

	if profileID == "" {
		for k := range Config.Profiles {
			viper.Set("profile", k)
			profileID = k
			break
		}
	}
	cp, ok := (Config.Profiles[profileID])
	if !ok {
		return ProfileConfig{}, fmt.Errorf("error loading profile '%s': no profile exists by this name in the config file", profileID)
	}
	return cp, nil
}

func findProjectFile() (string, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for prev := ""; dir != prev; dir = filepath.Dir(dir) {
		file := filepath.Join(dir, ConfigFileName)

		if _, err := os.Stat(file); err == nil {
			return file, nil
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

	err = yaml.Unmarshal(data, &Config.ProjectConfig)
	if err != nil {
		return err
	}
	return nil
}

func LoadProfileConfig() error {
	path, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("could not find current users home directory: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could read find the direktiv-configuration-file: %w", err)
	}

	err = yaml.Unmarshal(data, &Config.Profiles)
	if err != nil {
		return err
	}
	return nil
}

func GetConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find current users home directory: %w", err)
	}
	return filepath.Join(home, DefaultProfileConfigPath, DefaultProfileConfigName), nil
}

func getAddr() string {
	addr := viper.GetString("addr")
	return addr
}

func GetNamespace() string {
	namespace := viper.GetString("namespace")
	return namespace
}

func getInsecure() bool {
	return viper.GetBool("insecure")
}

func GetTLSConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: getInsecure()}
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

func GetRelativePath(configPath, targpath string) (string, error) {
	var err error

	if !filepath.IsAbs(configPath) {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return "", err
		}
	}

	if !filepath.IsAbs(targpath) {
		targpath, err = filepath.Abs(targpath)
		if err != nil {
			return "", err
		}
	}

	s, err := filepath.Rel(configPath, targpath)
	if err != nil {
		return "", err
	}

	path := filepath.ToSlash(s)
	path = strings.Trim(path, "/")

	return path, nil
}

func GetProjectFileLocation() string {
	projectFilePath := viper.GetString("projectFile")
	if projectFilePath != "" {
		projectFilePath = filepath.Dir(projectFilePath)
		projectFilePath = strings.TrimSuffix(projectFilePath, "/")
		return projectFilePath
	}
	return "."
}
