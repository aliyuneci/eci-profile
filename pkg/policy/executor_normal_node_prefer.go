package policy

import (
	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type NormalNodePreferExecutor struct {
}

func NewNormalNodePreferExecutor() Executor {
	return &NormalNodePreferExecutor{}
}

func (e *NormalNodePreferExecutor) OnPodCreating(selector *eciv1.Selector, pod *v1.Pod) ([]PatchInfo, error) {
	return nil, nil
}

func (e *NormalNodePreferExecutor) OnPodUnscheduled(selector *eciv1.Selector, pod *v1.Pod) (*utils.PatchOption, error) {
	if existVirtualTolerations(pod.Spec.Tolerations) {
		return nil, nil
	}
	patchOption := utils.NewPatchOption()
	tolerations := append(pod.Spec.Tolerations, virtualNodeToleration)
	patchOption.WithTolerations(tolerations)
	return patchOption, nil
}

func (e *NormalNodePreferExecutor) OnPodScheduled(selector *eciv1.Selector, pod *v1.Pod) (*utils.PatchOption, error) {
	patchOption := utils.NewPatchOption()
	if !existVirtualTolerations(pod.Spec.Tolerations) {
		tolerations := append(pod.Spec.Tolerations, virtualNodeToleration)
		patchOption.WithTolerations(tolerations)
	}
	patchOption.WithAnnotations(selector.Spec.Effect.Annotations).WithLabels(selector.Spec.Effect.Labels)
	return patchOption, nil
}
