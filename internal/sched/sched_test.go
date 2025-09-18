package sched

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	tclock "k8s.io/utils/clock/testing"
)

type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

type pub struct {
	subj string
	data []byte
	opts []nats.PubOpt
}
type sub struct {
	subj string
	cb   nats.MsgHandler
	opts []nats.SubOpt
}

type fakeJS struct {
	pubs []pub
	subs []sub
}

func (f *fakeJS) Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	f.pubs = append(f.pubs, pub{subj, append([]byte{}, data...), opts})
	return &nats.PubAck{Stream: "x", Sequence: 1}, nil
}
func (f *fakeJS) Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	f.subs = append(f.subs, sub{subj, cb, opts})
	return &nats.Subscription{}, nil
}

func TestFingerprintDoesNotMutateRule(t *testing.T) {
	r := &Rule{Namespace: "ns", WorkflowPath: "/wf", CreatedAt: time.Unix(10, 0), UpdatedAt: time.Unix(20, 0)}
	jason1, _ := json.Marshal(r)
	_ = r.Fingerprint()
	jason2, _ := json.Marshal(r)

	require.Equal(t, jason1, jason2, "Fingerprint must not zero CreatedAt")
}

func TestDispatchIfDue_PublishesTaskAndAdvancesRule(t *testing.T) {
	js := &fakeJS{}
	start := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	clk := tclock.NewFakeClock(start)
	s := New(js, clk, slog.New(slog.DiscardHandler))

	runAt := clk.Now()

	rule := &Rule{
		ID:           "rid",
		Namespace:    "ns",
		WorkflowPath: "/a",
		RunAt:        runAt,            // due
		CronExpr:     "*/30 * * * * *", // seconds
		Sequence:     5,
		CreatedAt:    clk.Now().Add(-time.Minute),
		UpdatedAt:    clk.Now().Add(-time.Minute),
	}

	err := s.dispatchIfDue(rule)
	require.NoError(t, err)

	// two publishes (task + rule update)
	require.Len(t, js.pubs, 2)

	// Task payload
	var task Task
	require.NoError(t, json.Unmarshal(js.pubs[0].data, &task))
	require.Equal(t, "ns", task.Namespace)
	require.Equal(t, "/a", task.WorkflowPath)
	require.WithinDuration(t, runAt, task.RunAt, 1*time.Second)

	// Rule update payload
	var updated Rule
	require.NoError(t, json.Unmarshal(js.pubs[1].data, &updated))
	require.Equal(t, "rid", updated.ID)
	require.Equal(t, rule.Namespace, updated.Namespace)
	require.WithinDuration(t, runAt.Add(30*time.Second), updated.RunAt, 0) // advanced
	require.True(t, updated.UpdatedAt.After(updated.CreatedAt))
}

func TestDispatchIfDue_SkipsWhenNotDue(t *testing.T) {
	js := &fakeJS{}
	start := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	clk := tclock.NewFakeClock(start)
	s := New(js, clk, slog.New(slog.DiscardHandler))

	rule := &Rule{
		ID:           "rid",
		Namespace:    "ns",
		WorkflowPath: "/a",
		RunAt:        clk.Now().Add(+5 * time.Second), // future => not due
		CronExpr:     "*/10 * * * * *",
	}

	err := s.dispatchIfDue(rule)
	require.Error(t, err) // "not-due"
	require.Len(t, js.pubs, 0, "no publishes when not due")
}
