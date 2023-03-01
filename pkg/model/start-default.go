package model

type DefaultStart struct {
	StartCommon `yaml:",inline"`
}

func (o *DefaultStart) GetEvents() []StartEventDefinition {
	return make([]StartEventDefinition, 0)
}

func (o *DefaultStart) Validate() error {
	if o == nil {
		return nil
	}

	if err := o.commonValidate(); err != nil {
		return err
	}

	return nil
}
