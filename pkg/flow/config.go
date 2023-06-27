package flow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/core"

	"github.com/jinzhu/copier"
)

type Config struct {
	Broadcast *ConfigBroadcast `json:"broadcast"`
}

type ConfigBroadcast struct {
	WorkflowCreate          *bool `json:"workflow.create"`
	WorkflowUpdate          *bool `json:"workflow.update"`
	WorkflowDelete          *bool `json:"workflow.delete"`
	DirectoryCreate         *bool `json:"directory.create"`
	DirectoryDelete         *bool `json:"directory.delete"`
	WorkflowVariableCreate  *bool `json:"workflow.variable.create"`
	WorkflowVariableUpdate  *bool `json:"workflow.variable.update"`
	WorkflowVariableDelete  *bool `json:"workflow.variable.delete"`
	NamespaceVariableCreate *bool `json:"namespace.variable.create"`
	NamespaceVariableUpdate *bool `json:"namespace.variable.update"`
	NamespaceVariableDelete *bool `json:"namespace.variable.delete"`
	InstanceVariableCreate  *bool `json:"instance.variable.create"`
	InstanceVariableUpdate  *bool `json:"instance.variable.update"`
	InstanceVariableDelete  *bool `json:"instance.variable.delete"`
	InstanceStarted         *bool `json:"instance.started"`
	InstanceSuccess         *bool `json:"instance.success"`
	InstanceFailed          *bool `json:"instance.failed"`
}

var defaultNamespaceConfig Config

func init() {
	err := json.Unmarshal([]byte(core.DefaultNamespaceConfig), &defaultNamespaceConfig)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal defaultNamespaceConfig: %w", err))
	}
}

// mergeIntoNamespaceConfig : Unmarshal sourceCfg and merge it's content onto self
// merged contents are then marshaled and returned.
func (c *Config) mergeIntoNamespaceConfig(sourceCfg []byte) ([]byte, error) {
	var sourceConfig Config

	err := json.Unmarshal(sourceCfg, &sourceConfig)
	if err != nil {
		return nil, err
	}

	err = copier.CopyWithOption(&sourceConfig, c, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return nil, fmt.Errorf("failed to merge config: %w", err)
	}

	mergedCfg, err := json.Marshal(sourceConfig)

	return mergedCfg, err
}

// loadNSConfig : loads config object from json cfgData.
func loadNSConfig(cfgData []byte) (*Config, error) {
	nsCFG := new(Config)
	dec := json.NewDecoder(bytes.NewReader(cfgData))
	dec.DisallowUnknownFields()

	err := dec.Decode(&nsCFG)

	return nsCFG, err
}

// broadcastEnabled : Checks if a broadcastTarget is available.
// If available return if it's value; mising target keys are returned as false.
func (c *Config) broadcastEnabled(broadcastTarget string) bool {
	var broadcastMap map[string]bool
	bData, err := json.Marshal(c.Broadcast)
	if err != nil {
		return false
	}

	err = json.Unmarshal(bData, &broadcastMap)
	if err != nil {
		return false
	}

	if enabled, ok := broadcastMap[broadcastTarget]; ok {
		return enabled
	}

	return false
}
