package policy

import (
	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type PatchInfo struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Executor interface {
	OnPodCreating(selector *eciv1.Selector, pod *v1.Pod) ([]PatchInfo, error)
	OnPodUnscheduled(selector *eciv1.Selector, pod *v1.Pod) (*utils.PatchOption, error)
}
