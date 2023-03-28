package policy

import (
	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type FairExecutor struct{}

func NewFairExecutor() Executor {
	return &FairExecutor{}
}

func (e *FairExecutor) OnPodCreating(selector *eciv1.Selector, pod *v1.Pod) ([]PatchInfo, error) {
	var patchInfos []PatchInfo
	if !existVirtualTolerations(pod.Spec.Tolerations) {
		patchInfos = append(patchInfos, addVirtualNodeToleration(pod))
	}
	if len(selector.Spec.Effect.Annotations) > 0 {
		patchInfos = append(patchInfos, addAnnotations(selector, pod))
	}
	if len(selector.Spec.Effect.Labels) > 0 {
		patchInfos = append(patchInfos, addLabels(selector, pod))
	}
	return patchInfos, nil
}

func (e *FairExecutor) OnPodUnscheduled(selector *eciv1.Selector, pod *v1.Pod) (*utils.PatchOption, error) {
	if existVirtualTolerations(pod.Spec.Tolerations) {
		return nil, nil
	}
	patchOption := utils.NewPatchOption()
	tolerations := append(pod.Spec.Tolerations, virtualNodeToleration)
	patchOption.WithTolerations(tolerations).
		WithAnnotations(selector.Spec.Effect.Annotations).
		WithLabels(selector.Spec.Effect.Labels)
	return patchOption, nil
}
