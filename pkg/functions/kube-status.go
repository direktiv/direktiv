package functions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	watcherTimeout = 60 * time.Minute
)

func (is *functionsServer) WatchFunctions(in *igrpc.WatchFunctionsRequest, out igrpc.FunctionsService_WatchFunctionsServer) error {
	cs, err := fetchServiceAPI()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %w", err)
	}

	annotations := in.GetAnnotations()
	labels := labels.Set(annotations).String()

	for {
		if done, err := is.watcherFunctions(cs, labels, out); err != nil {
			logger.Errorf("function watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("function watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}
}

func (is *functionsServer) watcherFunctions(cs *versioned.Clientset, labels string, out igrpc.FunctionsService_WatchFunctionsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())

	watch, err := cs.ServingV1().Services(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %w", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			s, ok := event.Object.(*v1.Service)
			if !ok {
				return false, nil
			}

			status, conds := statusFromCondition(s.Status.Conditions)
			resp := igrpc.WatchFunctionsResponse{
				Event: (*string)(&event.Type),
				Function: &igrpc.FunctionsInfo{
					Info:        serviceBaseInfo(s),
					Status:      &status,
					Conditions:  conds,
					ServiceName: &s.Name,
				},
			}

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %w", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("function watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
}

func (is *functionsServer) WatchRevisions(in *igrpc.WatchRevisionsRequest, out igrpc.FunctionsService_WatchRevisionsServer) error {
	var revisionFilter string

	if in.GetServiceName() == "" {
		return fmt.Errorf("service name can not be nil")
	}

	cs, err := fetchServiceAPI()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %w", err)
	}

	l := map[string]string{
		ServiceKnativeHeaderName: SanitizeLabel(in.GetServiceName()),
		// ServiceHeaderScope:       in.GetScope(),
	}

	if in.GetRevisionName() != "" {
		revisionFilter = SanitizeLabel(in.GetRevisionName())
	}

	labels := labels.Set(l).String()

	for {
		if done, err := is.watcherRevisions(cs, labels, revisionFilter, out); err != nil {
			logger.Errorf("revision watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("revision watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}
}

func (is *functionsServer) watcherRevisions(cs *versioned.Clientset, labels string, revisionFilter string, out igrpc.FunctionsService_WatchRevisionsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())

	watch, err := cs.ServingV1().Revisions(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %w", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():

			rev, ok := event.Object.(*v1.Revision)
			if !ok {
				return false, nil
			} else if revisionFilter != "" && rev.Name != revisionFilter {
				continue // skip
			}
			info := &igrpc.Revision{}

			// size and scale
			var sz, scale int32
			var gen int64
			fmt.Sscan(rev.Annotations[ServiceHeaderSize], &sz)
			fmt.Sscan(rev.Annotations["autoscaling.knative.dev/minScale"], &scale)
			fmt.Sscan(rev.Labels[ServiceTemplateGeneration], &gen)

			info.Size = &sz
			info.MinScale = &scale
			info.Generation = &gen

			// set status
			status, conds := statusFromCondition(rev.Status.Conditions)
			info.Status = &status
			info.Conditions = conds

			img, cmd := containerFromList(rev.Spec.Containers)
			info.Image = &img
			info.Cmd = &cmd

			// name
			svn := rev.Name
			info.Name = &svn

			ss := strings.Split(rev.Name, "-")
			info.Rev = &ss[len(ss)-1]

			info.ActualReplicas = int64(0)
			info.DesiredReplicas = int64(0)

			// replicas
			if rev.Status.ActualReplicas != nil {
				info.ActualReplicas = int64(*rev.Status.ActualReplicas)
			}

			if rev.Status.DesiredReplicas != nil {
				info.DesiredReplicas = int64(*rev.Status.DesiredReplicas)
			}

			// creation date
			t := rev.CreationTimestamp.Unix()
			info.Created = &t

			resp := igrpc.WatchRevisionsResponse{
				Event:    (*string)(&event.Type),
				Revision: info,
			}

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %w", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("revision watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
}

func (is *functionsServer) WatchPods(in *igrpc.WatchPodsRequest, out igrpc.FunctionsService_WatchPodsServer) error {
	if in.GetServiceName() == "" {
		return fmt.Errorf("service name can not be nil")
	}

	cs, err := getClientSet()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %w", err)
	}

	l := map[string]string{
		ServiceKnativeHeaderName: SanitizeLabel(in.GetServiceName()),
	}

	if in.GetRevisionName() != "" {
		l[ServiceKnativeHeaderRevision] = SanitizeLabel(in.GetRevisionName())
	}

	labels := labels.Set(l).String()

	for {
		if done, err := is.watcherPods(cs, labels, out); err != nil {
			logger.Errorf("pod watcher channel failed to restart: %s", err.Error())
			return err
		} else if done {
			// connection has ended
			return nil
		}
		logger.Debugf("pod watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}
}

func (is *functionsServer) watcherPods(cs *kubernetes.Clientset, labels string, out igrpc.FunctionsService_WatchPodsServer) (bool, error) {
	timeout := int64(watcherTimeout.Seconds())

	watch, err := cs.CoreV1().Pods(functionsConfig.Namespace).Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels,
		TimeoutSeconds: &timeout,
	})
	if err != nil {
		return false, fmt.Errorf("could start watcher: %w", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			p, ok := event.Object.(*corev1.Pod)
			if !ok {
				return false, nil
			}

			svc := p.Labels[ServiceKnativeHeaderName]
			srev := p.Labels[ServiceKnativeHeaderRevision]

			pod := igrpc.PodsInfo{
				Name:            &p.Name,
				Status:          (*string)(&p.Status.Phase),
				ServiceName:     &svc,
				ServiceRevision: &srev,
			}

			resp := igrpc.WatchPodsResponse{
				Event: (*string)(&event.Type),
				Pod:   &pod,
			}

			err = out.Send(&resp)
			if err != nil {
				return false, fmt.Errorf("failed to send event: %w", err)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		case <-out.Context().Done():
			logger.Debug("pod watcher server event connection closed")
			watch.Stop()
			return true, nil
		}
	}
}

func (is *functionsServer) WatchLogs(in *igrpc.WatchLogsRequest, out igrpc.FunctionsService_WatchLogsServer) error {
	if in.GetPodName() == "" {
		return fmt.Errorf("pod name can not be nil")
	}

	cs, err := getClientSet()
	if err != nil {
		return fmt.Errorf("could not create fetch client: %w", err)
	}

	req := cs.CoreV1().Pods(functionsConfig.Namespace).GetLogs(*in.PodName, &corev1.PodLogOptions{
		Container: "direktiv-container",
		Follow:    true,
	})

	plogs, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("could not get logs: %w", err)
	}
	defer plogs.Close()

	var done bool

	// Make sure stream is closed if client disconnects
	go func() {
		<-out.Context().Done()
		plogs.Close()
		done = true
	}()

	for {
		if done {
			break
		}
		buf := make([]byte, 2000)
		numBytes, err := plogs.Read(buf)
		if numBytes == 0 {
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		message := string(buf[:numBytes])
		resp := igrpc.WatchLogsResponse{
			Data: &message,
		}

		err = out.Send(&resp)
		if err != nil {
			return fmt.Errorf("log watcher failed to send event: %w", err)
		}
	}

	return nil
}
