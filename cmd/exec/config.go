package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// autoConfigPathFinder : Walk through parent directories of local workflow yaml until config file is found.
func autoConfigPathFinder() {
	var pathArg string
	var currentDir string

	// Find path is cmd args
	for i := 2; i < len(os.Args); i++ {
		if !strings.HasPrefix(os.Args[i-1], "-") && !strings.HasPrefix(os.Args[i], "-") {
			pathArg = os.Args[i]
			break
		}
	}

	wfWD, err := filepath.Abs(pathArg)
	if err != nil {
		log.Fatalf("Failed to locate workflow file in filesystem: %v\n", err)
	}

	if fStat, err := os.Stat(wfWD); err == nil && fStat.IsDir() {
		currentDir = wfWD
	} else {
		// Get parent dir if target local path is a file
		currentDir = filepath.Dir(wfWD)
	}

	for previousDir := ""; currentDir != previousDir; currentDir = filepath.Dir(currentDir) {
		cfgPath := filepath.Join(currentDir, DefaultConfigName)
		if _, err := os.Stat(cfgPath); err == nil {
			configPath = cfgPath
			configPathFromFlag = false
			break
		}
		previousDir = currentDir
	}
}

// Manually load config flag
func loadCfgFlag() {
	// flag.Parse()
	var foundFlag bool
	for _, arg := range os.Args {
		if foundFlag {
			configPath = arg
			break
		}

		if arg == "--config" || arg == "-c" {
			foundFlag = true
			continue
		}

		if strings.HasPrefix(arg, "-c=") {
			configPath = strings.TrimPrefix(arg, "-c=")
			break
		}

		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
			break
		}
	}
}

// configFlagHelpTextLoader : Generate suffix for flag help text to show set config value.
func configFlagHelpTextLoader(configKey string, sensitive bool) (flagHelpText string) {
	configValue := viper.GetString(configKey)

	if configValue != "" {
		if sensitive {
			flagHelpText = "(config \"***************\")"
		} else {
			flagHelpText = fmt.Sprintf("(config \"%s\")", configValue)
		}
	}

	return
}

//	configBindFlag : Binds cli flag for config value. If flag value is set, will be used instead of config value.
//	If config value is not set, mark flag as required.
func configBindFlag(cmd *cobra.Command, configKey string, required bool) {
	viper.BindPFlag(configKey, cmd.Flags().Lookup(configKey))
	if required && viper.GetString(configKey) == "" {
		cmd.MarkFlagRequired(configKey)
	}
}
