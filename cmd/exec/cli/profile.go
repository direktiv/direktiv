package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type profile struct {
	Address   string `json:"address"`
	Insecure  bool   `json:"insecure"`
	Namespace string `json:"namespace"`
	Token     string `json:"token"`
}

var configCmd = &cobra.Command{
	Use:   "profile",
	Short: "Add, list delete access profiles",
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(listProfilesCmd)
	configCmd.AddCommand(confAddCmd)
	configCmd.AddCommand(deleteProfileCmd)
}

var deleteProfileCmd = &cobra.Command{
	Use:   "delete [name of profile]",
	Args:  cobra.ExactArgs(1),
	Short: "Deletes profile with provided name",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := readConfig()
		if err != nil {
			return err
		}

		cmd.Printf("config file: %s\n", viper.ConfigFileUsed())

		configMap := viper.AllSettings()

		_, ok := configMap[args[0]]
		if !ok {
			return fmt.Errorf("profile %s does not exist", args[0])
		}

		delete(configMap, args[0])
		encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
		err = viper.ReadConfig(bytes.NewReader(encodedConfig))
		if err != nil {
			return err
		}
		err = viper.WriteConfig()
		if err != nil {
			return err
		}

		cmd.Printf("profile %s deleted\n", args[0])
		return nil
	},
}

var listProfilesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles in the profile file",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := readConfig()
		if err != nil {
			return err
		}

		cmd.Printf("config file: %s\n", viper.ConfigFileUsed())

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Profile", "URL", "Namespace", "Insecure", "Token"})

		for k, v := range viper.AllSettings() {
			m, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("wrong dataset for profiles")
			}

			token := fmt.Sprintf("%v", m["token"])
			if token != "" {
				token = "yes"
			} else {
				token = "none"
			}

			data := []string{k, fmt.Sprintf("%v", m["address"]), fmt.Sprintf("%v", m["namespace"]), fmt.Sprintf("%v", m["insecure"]), token}

			table.Append(data)
		}

		table.Render()

		return nil
	},
}

var confAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a profile configuration.",
	Long: `Add a profile to a existing profile list use the global flags to populate the config values. If the profile exists it will get overwritten.

Examples: 
  add -p myserver --address 192.168.122.232 -n mynamespace
  add -p myserver --address 192.168.122.232 -n mynamespace -t myaccesstoken --insecure`,

	RunE: func(cmd *cobra.Command, args []string) error {
		err := readConfig()
		if err != nil {
			return err
		}
		bindArgs()

		cmd.Printf("config file: %s\n", viper.ConfigFileUsed())

		name := RootCmd.PersistentFlags().Lookup("profile").Value.String()
		if name == "" {
			return fmt.Errorf("profile name required")
		}

		addr := viper.GetString(fmt.Sprintf("%s.address", name))

		if addr == "" {
			return fmt.Errorf("profile address required")
		}

		if !strings.HasPrefix(addr, "http") {
			return fmt.Errorf("address has to start with http:// ort https://")
		}

		if viper.GetString(fmt.Sprintf("%s.namespace", name)) == "" {
			return fmt.Errorf("profile namespace required")
		}

		err = viper.WriteConfig()
		if err != nil {
			return err
		}

		cmd.Printf("profile %s created or updated\n", name)
		return nil
	},
}
