package direktiv

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (is *ingressServer) CancelWorkflowInstance(ctx context.Context, in *ingress.CancelWorkflowInstanceRequest) (*emptypb.Empty, error) {

	err := is.wfServer.engine.hardCancelInstance(in.GetId(), "direktiv.cancels.api", "cancelled by api request")
	if err != nil {
		log.Errorf("error cancelling instance: %v", err)
	}
	return &emptypb.Empty{}, nil

}

func (is *ingressServer) GetWorkflowInstance(ctx context.Context, in *ingress.GetWorkflowInstanceRequest) (*ingress.GetWorkflowInstanceResponse, error) {

	var resp ingress.GetWorkflowInstanceResponse

	id := in.GetId()

	inst, err := is.wfServer.dbManager.getWorkflowInstance(ctx, id)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", id)
	}

	rev := int32(inst.Revision)

	var invokedBy string
	if wfID, err := inst.QueryWorkflow().FirstID(ctx); err == nil {
		invokedBy = wfID.String()
	} else {
		return nil, grpcDatabaseError(err, "workflow instance", id)
	}

	resp.Id = &id
	resp.Status = &inst.Status
	resp.InvokedBy = &invokedBy
	resp.BeginTime = timestamppb.New(inst.BeginTime)
	resp.Revision = &rev

	zt := time.Time{}
	if inst.EndTime != zt {
		resp.EndTime = timestamppb.New(inst.EndTime)
	}

	resp.Flow = inst.Flow
	resp.Input = []byte(inst.Input)
	resp.Output = []byte(inst.Output)

	resp.ErrorCode = &inst.ErrorCode
	resp.ErrorMessage = &inst.ErrorMessage

	return &resp, nil

}

func (is *ingressServer) GetWorkflowInstanceLogs(ctx context.Context, in *ingress.GetWorkflowInstanceLogsRequest) (*ingress.GetWorkflowInstanceLogsResponse, error) {

	var resp ingress.GetWorkflowInstanceLogsResponse

	instance := in.GetInstanceId()
	offset := in.GetOffset()
	limit := in.GetLimit()

	logs, err := is.wfServer.instanceLogger.QueryLogs(ctx, instance, int(limit), int(offset))
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", instance)
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for i := range logs.Logs {

		l := &logs.Logs[i]

		resp.WorkflowInstanceLogs = append(resp.WorkflowInstanceLogs, &ingress.GetWorkflowInstanceLogsResponse_WorkflowInstanceLog{
			Timestamp: timestamppb.New(time.Unix(0, l.Timestamp)),
			Message:   &l.Message,
			Context:   l.Context,
		})

	}

	return &resp, nil

}

func (is *ingressServer) WatchWorkflowInstanceLogs(in *ingress.WatchWorkflowInstanceLogsRequest, out ingress.DirektivIngress_WatchWorkflowInstanceLogsServer) error {

	logChannel, err := is.wfServer.instanceLogger.StreamLogs(context.Background(), in.GetInstanceId())
	if err != nil {
		return err
	}

	for {
		select {
		case <-out.Context().Done():
			log.Debug("watcher server event connection closed")
			return nil
		case event := <-logChannel:
			l, ok := event.(dlog.LogEntry)
			if !ok {
				log.Error("EVENT IS NOT A LOG ENTRY")
				return fmt.Errorf("got event error")
			}

			resp := ingress.WatchWorkflowInstanceLogsResponse{
				Level:     &l.Level,
				Timestamp: timestamppb.New(time.Unix(0, l.Timestamp)),
				Context:   l.Context,
				Message:   &l.Message,
			}

			err = out.Send(&resp)
			if err != nil {
				return fmt.Errorf("failed to send event: %v", err)
			}
		}
	}

	return nil

}

func (is *ingressServer) GetInstancesByWorkflow(ctx context.Context, in *ingress.GetInstancesByWorkflowRequest) (*ingress.GetInstancesByWorkflowResponse, error) {

	var resp ingress.GetInstancesByWorkflowResponse

	namespace := in.GetNamespace()
	workflow := in.GetWorkflow()
	offset := in.GetOffset()
	limit := in.GetLimit()

	workflowUID, err := is.wfServer.dbManager.getWorkflowByName(ctx, namespace, workflow)
	if err != nil {
		return nil, err
	}

	instances, err := is.wfServer.dbManager.getWorkflowInstancesByWFID(ctx, workflowUID.ID, int(offset), int(limit))
	if err != nil {
		return nil, err
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, inst := range instances {

		resp.WorkflowInstances = append(resp.WorkflowInstances, &ingress.GetInstancesByWorkflowResponse_WorkflowInstance{
			Id:        &inst.InstanceID,
			BeginTime: timestamppb.New(inst.BeginTime),
			Status:    &inst.Status,
		})

	}

	return &resp, nil
}

func (is *ingressServer) GetWorkflowInstances(ctx context.Context, in *ingress.GetWorkflowInstancesRequest) (*ingress.GetWorkflowInstancesResponse, error) {

	var resp ingress.GetWorkflowInstancesResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	instances, err := is.wfServer.dbManager.getWorkflowInstances(ctx, namespace, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, inst := range instances {

		resp.WorkflowInstances = append(resp.WorkflowInstances, &ingress.GetWorkflowInstancesResponse_WorkflowInstance{
			Id:        &inst.InstanceID,
			BeginTime: timestamppb.New(inst.BeginTime),
			Status:    &inst.Status,
		})

	}

	return &resp, nil

}
