package bytedata

import (
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/google/uuid"
)

func TestConvertLogMsg(t *testing.T) {
	input := make([]*internallogger.LogMsgs, 0)
	input = append(input, &internallogger.LogMsgs{
		Oid:   uuid.New(),
		T:     time.Now(),
		Msg:   "test",
		Level: "info",
	})
	resp, err := ConvertLogMsgToGrpcLog(input)
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
