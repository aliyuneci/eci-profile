package profile

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func isUnscheduledPod(pod *v1.Pod) bool {
	if pod.Status.Conditions == nil {
		klog.V(5).Infof("skip the pod without conditions: %s/%s (%s)", pod.Namespace, pod.Name, pod.UID)
		return false
	}

	conditions := pod.Status.Conditions

	var podScheduledCondition v1.PodCondition
	for _, condition := range conditions {
		if condition.Type == v1.PodScheduled {
			podScheduledCondition = condition
		}
	}

	if &podScheduledCondition != nil && podScheduledCondition.Status == v1.ConditionFalse && podScheduledCondition.Reason == v1.PodReasonUnschedulable {
		return true
	}

	return false
}
