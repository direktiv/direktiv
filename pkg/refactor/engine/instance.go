package engine

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
)

type Instance struct {
	Instance      *instancestore.InstanceData
	TelemetryInfo *InstanceTelemetryInfo
	RuntimeInfo   *InstanceRuntimeInfo
	Settings      *InstanceSettings
	DescentInfo   *InstanceDescentInfo
}

func ParseInstanceData(idata *instancestore.InstanceData) (*Instance, error) {
	ti, err := LoadInstanceTelemetryInfo(idata.TelemetryInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse telemetry info: %w", err)
	}

	var ri *InstanceRuntimeInfo
	if len(idata.RuntimeInfo) > 0 {
		ri, err = LoadInstanceRuntimeInfo(idata.RuntimeInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse runtime info: %w", err)
		}
	}

	settings, err := LoadInstanceSettings(idata.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings info: %w", err)
	}

	di, err := LoadInstanceDescentInfo(idata.DescentInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse descent info: %w", err)
	}

	return &Instance{
		Instance:      idata,
		TelemetryInfo: ti,
		RuntimeInfo:   ri,
		DescentInfo:   di,
		Settings:      settings,
	}, nil
}
