package policy

import (
	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type NormalNodeOnlyExecutor struct {
}

func NewNormalNodeOnlyExecutor() Executor {
	return &NormalNodeOnlyExecutor{}
}

func (e *NormalNodeOnlyExecutor) OnPodCreating(selector *eciv1.Selector, pod *v1.Pod) ([]PatchInfo, error) {
	return nil, nil
}

func (e *NormalNodeOnlyExecutor) OnPodUnscheduled(selector *eciv1.Selector, pod *v1.Pod) (*utils.PatchOption, error) {
	return nil, nil
}
