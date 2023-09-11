package functions

import (
	"fmt"
	"regexp"
	"strings"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	regex = "^[a-z]([-a-z0-9]{0,62}[a-z0-9])?$"
)

func validateLabel(name string) error {
	matched, err := regexp.MatchString(regex, name)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("invalid service name (must conform to regex: '%s')", regex)
	}

	return nil
}

func serviceBaseInfo(s *v1.Service) *igrpc.BaseInfo {
	var sz, scale int32
	fmt.Sscan(s.Annotations[ServiceHeaderSize], &sz)
	fmt.Sscan(s.Annotations[ServiceHeaderScale], &scale)

	n := s.Labels[ServiceHeaderName]
	ns := s.Labels[ServiceHeaderNamespaceID]
	nsName := s.Labels[ServiceHeaderNamespaceName]
	wf := s.Labels[ServiceHeaderWorkflowID]
	path := s.Labels[ServiceHeaderPath]
	rev := s.Labels[ServiceHeaderRevision]
	img, cmd := containerFromList(s.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers)

	info := &igrpc.BaseInfo{
		Name:          &n,
		Namespace:     &ns,
		Workflow:      &wf,
		Size:          &sz,
		MinScale:      &scale,
		Image:         &img,
		Cmd:           &cmd,
		NamespaceName: &nsName,
		Path:          &path,
		Revision:      &rev,
	}

	return info
}

func statusFromCondition(conditions []apis.Condition) (string, []*igrpc.Condition) {
	// status and status messages
	status := string(corev1.ConditionUnknown)

	var condList []*igrpc.Condition

	for m := range conditions {
		cond := conditions[m]

		if cond.Type == v1.RevisionConditionReady {
			status = string(cond.Status)
		}

		ct := string(cond.Type)
		st := string(cond.Status)
		c := &igrpc.Condition{
			Name:    &ct,
			Status:  &st,
			Reason:  &cond.Reason,
			Message: &cond.Message,
		}
		condList = append(condList, c)
	}

	return status, condList
}

func containerFromList(containers []corev1.Container) (string, string) {
	var img, cmd string

	for a := range containers {
		c := containers[a]

		if c.Name == containerUser {
			img = c.Image
			cmd = strings.Join(c.Command, ", ")
		}
	}

	return img, cmd
}

func createVolumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	volumes = append(volumes, functionsConfig.extraVolumes...)

	return volumes
}
