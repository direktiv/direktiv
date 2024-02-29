package states

import (
	"errors"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
)

func noMemory(logic Logic) error {
	if logic.GetMemory() != nil {
		return derrors.NewInternalError(errors.New("got unexpected savedata"))
	}

	return nil
}

func scheduleOnce(logic Logic, wakedata []byte) error {
	err := noMemory(logic)
	if err != nil {
		return err
	}

	if len(wakedata) != 0 {
		return derrors.NewInternalError(errors.New("got unexpected wakedata"))
	}

	return nil
}

func scheduleTwice(logic Logic, wakedata []byte) (bool, error) {
	err := noMemory(logic)
	if err != nil {
		return false, err
	}

	if len(wakedata) == 0 {
		return true, nil
	}

	return false, nil
}

func scheduleTwiceConst(logic Logic, wakedata []byte, expect string) (bool, error) {
	err := noMemory(logic)
	if err != nil {
		return false, err
	}

	if len(wakedata) == 0 {
		return true, nil
	}

	if string(wakedata) != expect {
		return false, derrors.NewInternalError(errors.New("got unexpected wakedata"))
	}

	return false, nil
}
