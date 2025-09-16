package sched

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"syscall"
	"testing"
	"time"

	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	tclock "k8s.io/utils/clock/testing"
)

func TestScheduler_TestProcessDueRules(t *testing.T) {
	ctx := context.Background()

	connStr, err := intNats.NewTestNats(t)
	require.NoError(t, err)

	// Connect to NATS + JetStream
	nc, err := nats.Connect(connStr)
	require.NoError(t, err)
	defer nc.Drain()

	js, err := intNats.SetupJetStream(context.Background(), nc)
	require.NoError(t, err)

	start := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	clk := tclock.NewFakeClock(start)

	// Build scheduler (real JS, real cache)
	s := New(js, clk, slog.New(slog.DiscardHandler))

	// Start rule subscription (fills cache)
	require.NoError(t, s.startRuleSubscription(ctx))

	// Seed one rule message in the rule stream as if SetRule had been called.
	now := clk.Now().UTC()
	rule := &Rule{
		ID:           "r1",
		Namespace:    "ns",
		WorkflowPath: "/wf",
		RunAt:        now, // due
		CronExpr:     "*/30 * * * * *",
		CreatedAt:    now.Add(-time.Minute),
		UpdatedAt:    now.Add(-time.Minute),
	}
	b, _ := json.Marshal(rule)
	_, err = js.Publish(fmt.Sprintf(intNats.SubjSchedRule, rule.Namespace, rule.ID), b)
	require.NoError(t, err)

	// Allow the async subscription to land the rule in the in-memory cache.
	time.Sleep(250 * time.Millisecond)
	// Manually trigger one tick to avoid waiting for the background ticker
	require.NoError(t, s.processDueRules())

	// Read from task stream to verify a task was produced
	sub, err := js.PullSubscribe(fmt.Sprintf(intNats.SubjSchedTask, rule.Namespace, rule.ID), "g1",
		nats.PullMaxWaiting(1),
	)
	require.NoError(t, err)

	msgs, err := sub.Fetch(10, nats.MaxWait(1*time.Second))
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	var task Task
	require.NoError(t, json.Unmarshal(msgs[0].Data, &task))
	require.Equal(t, "ns", task.Namespace)
	require.Equal(t, "/wf", task.WorkflowPath)

	// Optionally verify the rule update was republished with advanced RunAt.
	subRule, err := js.PullSubscribe(fmt.Sprintf(intNats.SubjSchedRule, rule.Namespace, rule.ID), "g2",
		nats.PullMaxWaiting(1),
	)
	require.NoError(t, err)
	defer subRule.Drain()

	msgs, err = subRule.Fetch(10, nats.MaxWait(1*time.Second))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(msgs), 1)

	var updated Rule
	require.NoError(t, json.Unmarshal(msgs[0].Data, &updated))
	require.WithinDuration(t, updated.RunAt.Add(-30*time.Second), rule.RunAt, 0)
}

func TestScheduler_EndToEnd(t *testing.T) {
	connStr, err := intNats.NewTestNats(t)
	require.NoError(t, err)

	type schedConn struct {
		sched *Scheduler
		conn  *nats.Conn
	}
	var scheds []schedConn

	lc := lifecycle.New(context.Background(), syscall.SIGQUIT)
	start := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
	clk := tclock.NewFakeClock(start)
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			clk.Step(100 * time.Millisecond)
		}
		lc.Stop()
	}()

	for i := 0; i < 10; i++ {
		nc, err := nats.Connect(connStr)
		require.NoError(t, err)

		js, err := intNats.SetupJetStream(context.Background(), nc)
		require.NoError(t, err)

		sched := New(js, clk, slog.With("inst", i))
		scheds = append(scheds, schedConn{
			sched: sched,
			conn:  nc,
		})
		err = sched.Start(lc)
		require.NoError(t, err)
	}

	_, err = scheds[0].sched.SetRule(context.Background(), &Rule{
		Namespace:    "ns",
		WorkflowPath: "/wf1",
		RunAt:        clk.Now(), // due
		CronExpr:     "*/2 * * * * *",
		CreatedAt:    clk.Now().Add(-time.Minute),
		UpdatedAt:    clk.Now().Add(-time.Minute),
	})
	require.NoError(t, err)
	_, err = scheds[0].sched.SetRule(context.Background(), &Rule{
		Namespace:    "ns",
		WorkflowPath: "/wf2",
		RunAt:        clk.Now(), // due
		CronExpr:     "*/2 * * * * *",
		CreatedAt:    clk.Now().Add(-time.Minute),
		UpdatedAt:    clk.Now().Add(-time.Minute),
	})
	require.NoError(t, err)

	<-lc.Done()
	t.Log("test terminated")

	err = lc.Wait(time.Second * 10)
	if err != nil {
		t.Log("harsh test termination")
	} else {
		t.Log("graceful test termination")
	}

	for _, sched := range scheds {
		err = sched.conn.Drain()
		require.NoError(t, err)
	}

	nc, err := nats.Connect(connStr)
	require.NoError(t, err)

	js, err := intNats.SetupJetStream(context.Background(), nc)
	require.NoError(t, err)
	// Read from task stream to verify a task was produced
	sub, err := js.PullSubscribe(fmt.Sprintf(intNats.SubjSchedTask, "*", "*"), "g1",
		nats.PullMaxWaiting(1),
	)
	require.NoError(t, err)

	msgs, err := sub.Fetch(100, nats.MaxWait(1*time.Second))
	require.NoError(t, err)

	var got []string
	for _, msg := range msgs {
		var task Task
		require.NoError(t, json.Unmarshal(msg.Data, &task))
		got = append(got, fmt.Sprintf("WorkflowPath:%s, CreatedAt:%s", task.WorkflowPath, task.CreatedAt.Format("15:04:05")))
	}

	want := []string{
		"WorkflowPath:/wf1, CreatedAt:01:00:00",
		"WorkflowPath:/wf2, CreatedAt:01:00:00",
		"WorkflowPath:/wf1, CreatedAt:01:00:02",
		"WorkflowPath:/wf2, CreatedAt:01:00:02",
		"WorkflowPath:/wf1, CreatedAt:01:00:04",
		"WorkflowPath:/wf2, CreatedAt:01:00:04",
		"WorkflowPath:/wf1, CreatedAt:01:00:06",
		"WorkflowPath:/wf2, CreatedAt:01:00:06",
		"WorkflowPath:/wf1, CreatedAt:01:00:08",
		"WorkflowPath:/wf2, CreatedAt:01:00:08",
		"WorkflowPath:/wf1, CreatedAt:01:00:10",
		"WorkflowPath:/wf2, CreatedAt:01:00:10",
	}
	slices.Sort(got)
	slices.Sort(want)

	require.Equal(t, want, got)

}
