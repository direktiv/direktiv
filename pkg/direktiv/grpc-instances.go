package direktiv

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (is *ingressServer) CancelWorkflowInstance(ctx context.Context, in *ingress.CancelWorkflowInstanceRequest) (*emptypb.Empty, error) {

	err := is.wfServer.engine.hardCancelInstance(in.GetId(), "direktiv.cancels.api", "cancelled by api request")
	if err != nil {
		appLog.Errorf("error cancelling instance: %v", err)
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

	lc := is.wfServer.components[util.LogComponent].(*logClient)
	r, err := lc.logsForInstance(instance, offset, limit)
	if err != nil {
		return nil, grpcDatabaseError(err, "instance", instance)
	}

	for i := range r {
		infoMap := r[i]

		// get msg
		msg := infoMap["msg"].(string)

		// get sec
		ts := infoMap["ts"].(float64)

		secs := int64(ts)
		nsecs := int64((ts - float64(secs)) * 1e9)
		tt := time.Unix(secs, nsecs)

		resp.WorkflowInstanceLogs = append(resp.WorkflowInstanceLogs, &ingress.GetWorkflowInstanceLogsResponse_WorkflowInstanceLog{
			Message:   &msg,
			Timestamp: timestamppb.New(tt.UTC()),
		})

	}

	return &resp, nil

}

func (is *ingressServer) WatchWorkflowInstanceLogs(in *ingress.WatchWorkflowInstanceLogsRequest, out ingress.DirektivIngress_WatchWorkflowInstanceLogsServer) error {
	instance := in.GetInstanceId()
	timeTracker := float64(0)

	pollTicker := time.NewTicker(250 * time.Millisecond)
	defer pollTicker.Stop()

	for {
		select {
		case <-out.Context().Done():
			pollTicker.Stop()
			return nil
		case <-pollTicker.C:

			lc := is.wfServer.components[util.LogComponent].(*logClient)
			r, err := lc.logsForInstanceAfterTime(instance, timeTracker)
			if err != nil {
				return grpcDatabaseError(err, "instance", instance)
			}

			for i := range r {
				infoMap := r[i]

				// get msg
				msg := infoMap["msg"].(string)
				msg = strings.TrimSuffix(msg, "\n")

				// get sec
				ts := infoMap["ts"].(float64)

				// get level
				level := infoMap["level"].(string)

				secs := int64(ts)
				nsecs := int64((ts - float64(secs)) * 1e9)
				tt := time.Unix(secs, nsecs)

				resp := ingress.WatchWorkflowInstanceLogsResponse{
					Level:     &level,
					Timestamp: timestamppb.New(tt.UTC()),
					// Context:   l.Context,
					Message: &msg,
				}

				err = out.Send(&resp)
				if err != nil {
					pollTicker.Stop()
					return fmt.Errorf("failed to send event: %v", err)
				}

				// update time tracker
				if i == len(r)-1 {
					timeTracker = ts
				}
			}

		}
	}
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
