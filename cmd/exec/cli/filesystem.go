package cli

import (
	"fmt"
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

		// at this stage there is a direktivignore file
		// matcher, err := loadIgnoresMatcher(filepath.Join(projectRoot, ".direktivignore"))
		// if err != nil {
		// 	return err
		// }

		uploader, err := newUploader(projectRoot, p)
		if err != nil {
			return err
		}

		// fmt.Printf("PROJECT ROOT %v\n", projectRoot)
		// fmt.Printf("PATH         %v\n", fullPath)

		// fp, err := filepath.Rel(projectRoot, fullPath)
		// if err != nil {
		// 	return err
		// }

		// fmt.Printf("PATH CALC    %v\n", fp)
		// fp = fmt.Sprintf("/%s", fp)

		fmt.Println("------------------------")
		err = filepath.Walk(args[0], func(path string, info os.FileInfo, err error) error {

			fullPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			// http://192.168.0.108/api/v2/namespaces/test/files/
			// {
			// 	"name": "sss.yaml",
			// 	"data": "ZGlyZWt0aXZfYXBpOiB3b3JrZmxvdy92MQpkZXNjcmlwdGlvbjogQSBzaW1wbGUgJ25vLW9wJyBzdGF0ZSB0aGF0IHJldHVybnMgJ0hlbGxvIHdvcmxkIScKc3RhdGVzOgotIGlkOiBoZWxsb3dvcmxkCiAgdHlwZTogbm9vcAogIHRyYW5zZm9ybToKICAgIHJlc3VsdDogSGVsbG8gd29ybGQhCg==",
			// 	"type": "workflow",
			// 	"mimeType": "application/yaml"
			//   }

			// fmt.Printf("FULL %v\n", fullPath)

			// fp, err := filepath.Rel(projectRoot, fullPath)
			// if err != nil {
			// 	return err
			// }

			// fmt.Printf("PATH CALC    %v\n", fp)

			p, err := GetRelativePath(projectRoot, fullPath)
			if err != nil {
				return err
			}
			// fmt.Printf("!!! %v %v\n", p, err)

			// fmt.Printf("MATCHER %+v\n", matcher)

			if info.IsDir() {
				err = uploader.createDirectory(p)
			} else {
				err = uploader.createFile(p, fullPath)
			}

			if err != nil {
				return err
			}

			// if matcher.Match(strings.Split(p, string(os.PathSeparator)), info.IsDir()) {
			// 	fmt.Println("MATCH")
			// } else {
			// 	fmt.Println("NO MATCH")
			// }

			// println(info.Name())
			// fmt.Println(path)

			// fmt.Printf(">>> %v %v\n", fp, path)

			// // fmt.Println(strings.Split(path, fp))

			// fp, err := filepath.Rel(projectRoot, path)
			// if err != nil {
			// 	return err
			// }
			// fmt.Printf("PATH CALC    %v\n", fp)
			// p, e := GetRelativePath(fp, path)
			// fmt.Printf("FILES %v %v\n", p, e)
			// a, b := filepath.Rel(path, fp)
			// fmt.Printf("!! %v %v\n", a, b)
			// 	if strings.HasSuffix(info.Name(), ".direktiv.ts") {
			// 		// workflow typescript
			// 	} else if strings.HasSuffix(info.Name(), ".yaml" || )  {

			// 	}

			return nil
		})

		fmt.Println(err)

		// configMap := viper.AllSettings()

		// _, ok := configMap[args[0]]
		// if !ok {
		// 	return fmt.Errorf("profile %s does not exist", args[0])
		// }

		// delete(configMap, args[0])
		// encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
		// err = viper.ReadConfig(bytes.NewReader(encodedConfig))
		// if err != nil {
		// 	return err
		// }
		// err = viper.WriteConfig()
		// if err != nil {
		// 	return err
		// }

		// cmd.Printf("profile %s deleted\n", args[0])
		return nil
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
