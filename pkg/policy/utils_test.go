package policy

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestAddVirtualNodeToleration(t *testing.T) {
	pod := &v1.Pod{}
	patchInfo := addVirtualNodeToleration(pod)
	if patchInfo.Op != "add" {
		t.Fatalf("test add virtual node toleration failed, patchInfo's Op is %s", patchInfo.Op)
	}
	if patchInfo.Path != "/spec/tolerations" {
		t.Fatalf("test add virtual node toleration failed, patchInfo's Path is %s", patchInfo.Path)
	}
	tolerations, ok := patchInfo.Value.([]v1.Toleration)
	if !ok {
		t.Fatalf("test add virtual node toleration failed, patchInfo's Value is %v", patchInfo.Value)
	}
	var index int
	for i := range tolerations {
		if tolerations[i].Key == vnodeNodeSelectorKey {
			index = i
			break
		}
	}
	if tolerations[index].Key != vnodeNodeSelectorKey {
		t.Fatalf("test add virtual node toleration failed, toleration key is %s", tolerations[index].Key)
	}
	if tolerations[index].Value != vnodeNodeSelectorVal {
		t.Fatalf("test add virtual node toleration failed, toleration value is %s", tolerations[index].Value)
	}
	if tolerations[index].Operator != v1.TolerationOpEqual {
		t.Fatalf("test add virtual node toleration failed, toleration operator is %s", tolerations[index].Operator)
	}
	if tolerations[index].Effect != v1.TaintEffectNoSchedule {
		t.Fatalf("test add virtual node toleration failed, toleration effect is %s", tolerations[index].Effect)
	}
}
func TestAddVirtualNodeSelector(t *testing.T) {
	patchInfo := addVirtualNodeSelector()
	if patchInfo.Op != "replace" {
		t.Fatalf("test add virtual node selector failed, patchInfo's Op is %s", patchInfo.Op)
	}
	if patchInfo.Path != "/spec/nodeSelector" {
		t.Fatalf("test add virtual node selecetor failed, patchInfo's Path is %s", patchInfo.Path)
	}
	nodeSelector, ok := patchInfo.Value.(map[string]string)
	if !ok {
		t.Fatalf("test add virtual node selector failed, patchInfo's Value is %v", patchInfo.Value)
	}
	if val, ok := nodeSelector[vnodeNodeSelectorKey]; !ok || val != vnodeNodeSelectorVal {
		t.Fatalf("test add virtual node selector failed, nodeSelector is %v", nodeSelector)
	}
}
