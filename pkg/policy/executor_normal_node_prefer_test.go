package policy

import (
	"reflect"
	"testing"

	eciv1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNormalNodePreferOnPodUnscheduled(t *testing.T) {
	for desc, test := range map[string]struct {
		pod         *v1.Pod
		selector    *eciv1beta1.Selector
		mutatePodFn func(*v1.Pod)
		expect      *utils.PatchOption
		expectErr   error
	}{
		"test exist virtual node tolerations": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.PodSpec{},
			},
			mutatePodFn: func(pod *v1.Pod) {
				pod.Spec.Tolerations = []v1.Toleration{
					{
						Key:      vnodeNodeSelectorKey,
						Value:    vnodeNodeSelectorVal,
						Operator: v1.TolerationOpEqual,
						Effect:   v1.TaintEffectNoSchedule,
					},
				}
			},
			expect:    nil,
			expectErr: nil,
		},
		"test not exist virtual node tolerations": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.PodSpec{},
			},
			selector: &eciv1beta1.Selector{
				Spec: eciv1beta1.SelectorSpec{
					Effect: &eciv1beta1.SideEffect{},
				},
			},
			expect: &utils.PatchOption{
				Spec: struct {
					Tolerations []v1.Toleration "json:\"tolerations,omitempty\""
				}{
					[]v1.Toleration{
						{
							Key:      vnodeNodeSelectorKey,
							Value:    vnodeNodeSelectorVal,
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectErr: nil,
		},
		"test not exist virtual node tolerations and include annotations, labels": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.PodSpec{},
			},
			selector: &eciv1beta1.Selector{
				Spec: eciv1beta1.SelectorSpec{
					Effect: &eciv1beta1.SideEffect{
						Annotations: map[string]string{
							"foo": "boo",
						},
						Labels: map[string]string{
							"foo": "boo",
						},
					},
				},
			},
			expect: &utils.PatchOption{
				Metadata: struct {
					Annotations map[string]string "json:\"annotations,omitempty\""
					Labels      map[string]string "json:\"labels,omitempty\""
				}{
					Annotations: map[string]string{
						"foo": "boo",
					},
					Labels: map[string]string{
						"foo": "boo",
					},
				},
				Spec: struct {
					Tolerations []v1.Toleration "json:\"tolerations,omitempty\""
				}{
					[]v1.Toleration{
						{
							Key:      vnodeNodeSelectorKey,
							Value:    vnodeNodeSelectorVal,
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectErr: nil,
		},
	} {
		if test.mutatePodFn != nil {
			test.mutatePodFn(test.pod)
		}
		executor := NewNormalNodePreferExecutor()
		actual, err := executor.OnPodUnscheduled(test.selector, test.pod)
		if err != nil && err != test.expectErr {
			t.Fatalf("[%s] executor on pod unscheduled failed, err: %v", desc, err)
		}
		if !reflect.DeepEqual(actual, test.expect) {
			t.Fatalf("[%s] executor on pod unscheduled failed, actual: %v, expect %v", desc, actual, test.expect)
		}
	}
}
