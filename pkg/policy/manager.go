package policy

import (
	eciv1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"
	"eci.io/eci-profile/pkg/resource"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

const (
	ExecutorNameFair             = "Fair"
	ExecutorNameNormalNodeOnly   = "NormalNodeOnly"
	ExecutorNameNormalNodePrefer = "NormalNodePrefer"
	ExecutorNameVirtualNodeOnly  = "VirtualNodeOnly"
)

type Manager struct {
	executors map[string]Executor
}

func NewManager(rm *resource.Manager) *Manager {
	return &Manager{
		executors: map[string]Executor{
			ExecutorNameFair:             NewFairExecutor(),
			ExecutorNameNormalNodeOnly:   NewNormalNodeOnlyExecutor(),
			ExecutorNameNormalNodePrefer: NewNormalNodePreferExecutor(),
			ExecutorNameVirtualNodeOnly:  NewVirtualNodeOnlyExecutor(),
		},
	}
}

func (m *Manager) OnPodCreating(selector *eciv1beta1.Selector, pod *v1.Pod) ([]PatchInfo, error) {
	executor := m.findExecutor(selector)
	return executor.OnPodCreating(selector, pod)
}

func (m *Manager) OnPodUnscheduled(selector *eciv1beta1.Selector, pod *v1.Pod) (*utils.PatchOption, error) {
	executor := m.findExecutor(selector)
	return executor.OnPodUnscheduled(selector, pod)
}

func (m *Manager) findExecutor(selector *eciv1beta1.Selector) Executor {
	executorName := ExecutorNameVirtualNodeOnly
	policy := selector.Spec.Policy
	switch {
	case policy.Fair != nil:
		executorName = ExecutorNameFair
	case policy.VirtualNodeOnly != nil:
		executorName = ExecutorNameVirtualNodeOnly
	case policy.NormalNodeOnly != nil:
		executorName = ExecutorNameNormalNodeOnly
	case policy.NormalNodePrefer != nil:
		executorName = ExecutorNameNormalNodePrefer
	}
	return m.executors[executorName]
}
