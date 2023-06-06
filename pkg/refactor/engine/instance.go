package engine

import (
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
		return nil, err
	}

	ri, err := LoadInstanceRuntimeInfo(idata.RuntimeInfo)
	if err != nil {
		return nil, err
	}

	settings, err := LoadInstanceSettings(idata.Settings)
	if err != nil {
		return nil, err
	}

	di, err := LoadInstanceDescentInfo(idata.DescentInfo)
	if err != nil {
		return nil, err
	}

	return &Instance{
		Instance:      idata,
		TelemetryInfo: ti,
		RuntimeInfo:   ri,
		DescentInfo:   di,
		Settings:      settings,
	}, nil
}
