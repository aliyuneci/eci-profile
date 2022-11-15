package policy

import (
	eciv1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"
	v1 "k8s.io/api/core/v1"
)

const (
	vnodeNodeSelectorKey = "k8s.aliyun.com/vnode"
	vnodeNodeSelectorVal = "true"
)

var (
	virtualNodeToleration = v1.Toleration{
		Key:      vnodeNodeSelectorKey,
		Value:    vnodeNodeSelectorVal,
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}
)

func addVirtualNodeToleration(pod *v1.Pod) PatchInfo {
	tolerations := pod.Spec.Tolerations
	tolerations = append(tolerations, virtualNodeToleration)
	return PatchInfo{
		Op:    "add",
		Path:  "/spec/tolerations",
		Value: tolerations,
	}
}

func addVirtualNodeSelector() PatchInfo {
	return PatchInfo{
		Op:   "replace",
		Path: "/spec/nodeSelector",
		Value: map[string]string{
			vnodeNodeSelectorKey: vnodeNodeSelectorVal,
		},
	}
}

func addAnnotations(selector *eciv1beta1.Selector, pod *v1.Pod) PatchInfo {
	annotations := pod.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	for key, value := range selector.Spec.Effect.Annotations {
		annotations[key] = value
	}
	return PatchInfo{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: annotations,
	}
}

func addLabels(selector *eciv1beta1.Selector, pod *v1.Pod) PatchInfo {
	labels := pod.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	for key, value := range selector.Spec.Effect.Labels {
		labels[key] = value
	}

	return PatchInfo{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: labels,
	}
}

func existVirtualTolerations(tolerations []v1.Toleration) bool {
	for _, toleration := range tolerations {
		if toleration.Key == vnodeNodeSelectorKey &&
			toleration.Value == vnodeNodeSelectorVal &&
			toleration.Operator == v1.TolerationOpEqual &&
			toleration.Effect == v1.TaintEffectNoSchedule {
			return true
		}
	}

	return false
}
