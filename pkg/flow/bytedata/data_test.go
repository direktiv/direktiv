package bytedata

import (
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logengine"
)

func TestConvertLogMsgForOutput(t *testing.T) {
	input := make([]*logengine.LogEntry, 0)
	field := make(map[string]interface{})
	field["level"] = "info"
	input = append(input, &logengine.LogEntry{
		T:      time.Now().UTC(),
		Msg:    "test",
		Fields: field,
	})
	resp, err := ConvertLogMsgForOutput(input)
	if err != nil {
		t.Errorf("got unexpected error, %s", err)
	}
	if len(resp) != len(input) {
		t.Errorf("response has wrong length, should: %d, got : %d", len(input), len(resp))
	}
	res := resp[0].T.AsTime()
	expected := input[0].T
	if !res.Equal(expected) {
		t.Errorf("time is wrong; expected: %s, got : %s", expected, res)
	}
}
