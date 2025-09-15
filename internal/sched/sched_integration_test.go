package sched

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestScheduler_EndToEnd(t *testing.T) {
	ctx := context.Background()

	connStr, err := database.NewTestNats(t)
	require.NoError(t, err)

	// Connect to NATS + JetStream
	nc, err := nats.Connect(connStr)
	require.NoError(t, err)
	defer nc.Drain()

	js, err := intNats.SetupJetStream(context.Background(), nc)
	require.NoError(t, err)

	// Build scheduler (real JS, real cache)
	s := New(js)

	// Start rule subscription (fills cache)
	require.NoError(t, s.startRuleSubscription(ctx))

	// Seed one rule message in the rule stream as if SetRule had been called.
	now := time.Now().UTC()
	rule := &Rule{
		ID:           "r1",
		Namespace:    "ns",
		WorkflowPath: "/wf",
		RunAt:        now.Add(-1 * time.Second), // due
		CronExpr:     30,                        // seconds
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
