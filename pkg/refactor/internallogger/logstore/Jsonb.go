package logstore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		if err := json.Unmarshal(value.([]byte), &j); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(fmt.Sprint(v)), &j); err != nil {
			return err
		}
	}

	return nil
}
