package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "filesystem",
	Short: "Execute flows and push files",
}

func init() {
	RootCmd.AddCommand(instancesCmd)
	instancesCmd.AddCommand(instancesPushCmd)
}

var instancesPushCmd = &cobra.Command{
	Use:   "push [name of file/directory]",
	Args:  cobra.ExactArgs(1),
	Short: "Push files to Direktiv",
	RunE: func(cmd *cobra.Command, args []string) error {

		p, err := prepareCommand(cmd)
		if err != nil {
			return err
		}

		fullPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		projectRoot, err := findProjectRoot(fullPath)
		if err != nil {
			return err
		}

		uploader, err := newUploader(projectRoot, p)
		if err != nil {
			return err
		}

		err = filepath.Walk(args[0], func(path string, info os.FileInfo, err error) error {

			fullPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			p, err := GetRelativePath(projectRoot, fullPath)
			if err != nil {
				return err
			}

			if uploader.matcher.Match(strings.Split(p, string(os.PathSeparator)), info.IsDir()) {
				cmd.Printf("skipping object %s\n", p)
				return nil
			}

			if info.IsDir() {
				err = uploader.createDirectory(p)
			} else {
				err = uploader.createFile(p, fullPath)
			}

			if err != nil {
				cmd.Printf("error creating object %s: %s\n", p, err.Error())
			}

			return nil
		})

		return err
	},
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
