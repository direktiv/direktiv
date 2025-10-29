package runtime

import (
	"fmt"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/go-viper/mapstructure/v2"
	"github.com/grafana/sobek"
)

func (rt *Runtime) action(c map[string]any) sobek.Value {
	fmt.Println("HELLO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1")
	var config core.ActionConfig
	err := mapstructure.Decode(c, &config)
	if err != nil {
		fmt.Println(err)
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling action configuration: %s", err.Error())))
	}

	fmt.Println(config)

	actionFunc := func(payload any) sobek.Value {
		fmt.Println(payload)
		rt.onAction(config)

		return rt.vm.ToValue("return value")
	}

	return rt.vm.ToValue(actionFunc)
}
