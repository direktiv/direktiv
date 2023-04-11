package cmd

import (
	"bytes"
	"crypto/tls"
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

const DefaultConfigName = project.ConfigFile

type ProfileConfig struct {
	ID        string `yaml:"id" mapstructure:"profile"`
	Addr      string `yaml:"addr" mapstructure:"addr"`
	Path      string `yaml:"path" mapstructure:"path"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`
	Auth      string `yaml:"auth" mapstructure:"auth"`
	MaxSize   int64  `yaml:"max-size" mapstructure:"max-size"`
}

type ConfigFile struct {
	ProfileConfig  `yaml:",inline" mapstructure:",squash"`
	project.Config `yaml:",inline" mapstructure:",squash"`
	Profiles       []ProfileConfig `yaml:"profiles,flow" mapstructure:"profiles"`
	profile        string
	path           string
}

var (
	config   ConfigFile
	Globbers []glob.Glob
)

func loadConfig(cmd *cobra.Command) {
	chdir, err := cmd.Flags().GetString("directory")
	if err != nil {
		Fail("error loading 'directory' flag: %v", err)
	}

	if chdir != "" && chdir != "." {
		err = os.Chdir(chdir)
		if err != nil {
			Fail("error chanding directory: %v", err)
		}
		Printlog("changed to directory: %s", chdir)
	}

	path := findConfig()

	Globbers = make([]glob.Glob, 0)
	for idx, pattern := range config.Ignore {
		g, err := glob.Compile(pattern)
		if err != nil {
			Fail("Failed to parse %dth ignore pattern: %w", idx, err)
		}
		Globbers = append(Globbers, g)
	}

	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		Fail("error loading 'profile' flag: %v", err)
	}

	config.profile = profile
	var cp *ProfileConfig

	if config.profile != "" {
		for idx := range config.Profiles {
			if config.Profiles[idx].ID == config.profile {
				cp = &(config.Profiles[idx])
				break
			}
		}

		if cp == nil {
			Fail("error loading profile '%s': no profile exists by this name in the config file", config.profile)
		}
	} else if len(config.Profiles) > 0 {
		cp = &(config.Profiles[0])
	}

	if path != "" {
		config.path = path

		if cp == nil {
			cp = &config.ProfileConfig
		}

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
}

func findConfig() string {
	dir, err := filepath.Abs(".")
	if err != nil {
		Fail("Failed to locate place in filesystem: %v\n", err)
	}

	for prev := ""; dir != prev; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, DefaultConfigName)

		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				Fail("Failed to read config file: %v", err)
			}

			err = yaml.Unmarshal(data, &config)
			if err != nil {
				Fail("Failed to parse config file: %v", err)
			}

			if len(config.Profiles) > 0 {
				if config.Addr != "" || config.ID != "" || config.Auth != "" || config.MaxSize != 0 ||
					config.Namespace != "" || config.Path != "" {
					Fail("config file cannot have top-level values alongside profiles")
				}
			}

			return path
		}

		prev = dir
	}

	return ""
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
	path = strings.TrimSuffix(path, ".yaml")
	path = strings.TrimSuffix(path, ".yml")

	path = strings.Trim(path, "/")

	return path
}

func GetPath(targpath string) string {
	path := viper.GetString("path")

	if path != "" {
		fj := filepath.Join(path, filepath.Base(targpath))
		path = strings.Trim(fj, "/")
		return path
	}

	// if config file was found automatically, generate path relative to config dir

	configPath := GetConfigPath()

	return GetRelativePath(configPath, targpath)
}

func GetConfigPath() string {
	if config.path != "" {
		path := config.path
		path = filepath.Dir(path)
		path = strings.TrimSuffix(path, "/")
		return path
	}

	return "."
}
