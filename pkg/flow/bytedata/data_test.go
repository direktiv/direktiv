package bytedata

// import (
// 	"testing"
// 	"time"

// 	"github.com/direktiv/direktiv/pkg/flow/ent"
// 	"github.com/google/uuid"
// )

// func TestConvertLogMsgForOutput(t *testing.T) {
// 	input := make([]*ent.LogMsg, 0)
// 	input = append(input, &ent.LogMsg{
// 		ID:    uuid.New(),
// 		T:     time.Now(),
// 		Msg:   "test",
// 		Level: "info",
// 	})
// 	resp, err := ConvertLogMsgForOutput(input)
// 	if err != nil {
// 		t.Errorf("got unexpected error, %s", err)
// 	}
// 	if len(resp) != len(input) {
// 		t.Errorf("response has wrong length, should: %d, got : %d", len(input), len(resp))
// 	}
// 	res := resp[0].T.AsTime()
// 	expected := input[0].T
// 	if !res.Equal(expected) {
// 		t.Errorf("time is wrong; expected: %s, got : %s", expected, res)
// 	}
// }
