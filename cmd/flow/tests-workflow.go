package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
)

const simpleWorkflow = `
states:
  - type: noop
    id: a
`

const aCheckedCloudEvent = `
{
  "specversion": "1.0",
  "type": "%s-checked",  
  "source": "https://test.io",
  "id": "%v",
  "data": {
	  "event": "%s"
  }
}
`

const startTypeEventWorkflow = `
start:
  type: event
  state: a
  event:
    type: a-checked
states:
- type: noop
  id: a
`

const startTypeEventAndWorkflow = `
start:
  type: eventsAnd
  state: a
  events:
  - type: a-checked
  - type: b-checked
states:
- type: noop
  id: a
`

const startTypeEventXorWorkflow = `
start:
  type: eventsXor
  events:
  - type: a-checked
  - type: b-checked
states:
- type: noop
  id: a
`

const startTypeScheduledWorkflow = `
start:
  type: scheduled
  state: a
  cron: "* * * * *"
states:
- type: noop
  id: a
`

func testStartTypeEvent(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(startTypeEventWorkflow),
	})
	if err != nil {
		return err
	}

	// send broadcast event
	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "a", uuid.New().String(), "a")),
	})
	if err != nil {
		return err
	}

	// go is to fast
	time.Sleep(time.Second * 1)

	resp, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	instances := resp.GetInstances().GetResults()
	for _, instance := range instances {
		instanceNode := instance.GetAs()
		if instanceNode == "/testwf" || instanceNode == "testwf" {
			return nil
		}
	}

	return errors.New("instance was never created from start event trigger")
}

func testStartTypeEventAnd(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(startTypeEventAndWorkflow),
	})
	if err != nil {
		return err
	}

	// send broadcast event
	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "a", uuid.New().String(), "a")),
	})
	if err != nil {
		return err
	}

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "b", uuid.New().String(), "b")),
	})
	if err != nil {
		return err
	}

	// go is to fast
	time.Sleep(time.Second * 1)

	resp, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	instances := resp.GetInstances().GetResults()
	for _, instance := range instances {
		instanceNode := instance.GetAs()
		if instanceNode == "/testwf" || instanceNode == "testwf" {
			return nil
		}
	}

	return errors.New("instance was never created from start eventAnd trigger")
}

func testStartTypeEventXor(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(startTypeEventXorWorkflow),
	})
	if err != nil {
		return err
	}

	// send cloud events two instances should be created.
	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "a", uuid.New().String(), "a")),
	})
	if err != nil {
		return err
	}

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "b", uuid.New().String(), "b")),
	})
	if err != nil {
		return err
	}

	// go is to fast
	time.Sleep(time.Second * 1)

	resp, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	instances := resp.GetInstances().GetResults()

	a := false
	b := false

	for _, instance := range instances {
		instanceNode := instance.GetAs()
		if instanceNode == "/testwf" || instanceNode == "testwf" {
			// check for output
			output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
				Instance:  instance.GetId(),
				Namespace: namespace,
			})
			if err != nil {
				return err
			}
			d := output.GetData()
			var ini map[string]interface{}
			err = json.Unmarshal(d, &ini)
			if err != nil {
				return err
			}
			if ini["a-checked"] != nil {
				a = true
			}
			if ini["b-checked"] != nil {
				b = true
			}
		}
	}

	if a && b {
		return nil
	}

	return errors.New("one or two instances weren't created from start eventsXor trigger")
}

func testStartTypeCron(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(startTypeScheduledWorkflow),
	})
	if err != nil {
		return err
	}

	// sleep for 3 minutes and 20 seconds then check if 3 instances we're created
	time.Sleep(200 * time.Second)

	resp, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	instances := resp.GetInstances().GetResults()

	if len(instances) == 3 || len(instances) == 4 {
		return nil
	}

	return fmt.Errorf("cron job schedule instances created: %v wanted 3", len(instances))

}
