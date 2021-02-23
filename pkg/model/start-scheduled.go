package model

type ScheduledStart struct {
	StartCommon `yaml:",inline"`
	Cron        string `yaml:"cron,omitempty"`
}

func (o *ScheduledStart) GetEvents() []StartEventDefinition {
	return make([]StartEventDefinition, 0)
}

func (o *ScheduledStart) Validate() error {
	if o == nil {
		return nil
	}

	if err := o.commonValidate(); err != nil {
		return err
	}

	return nil
}
