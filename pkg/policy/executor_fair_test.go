package policy

import (
	"reflect"
	"testing"

	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFairOnPodCreatingWithTolerations(t *testing.T) {
	for desc, test := range map[string]struct {
		pod         *v1.Pod
		selector    *eciv1.Selector
		mutatePodFn func(*v1.Pod)
		expect      []PatchInfo
	}{
		"test not exist virtual node tolerations": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.PodSpec{},
			},
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{},
				},
			},
			expect: []PatchInfo{
				{
					Op:   "add",
					Path: "/spec/tolerations",
					Value: []v1.Toleration{
						{
							Key:      vnodeNodeSelectorKey,
							Value:    vnodeNodeSelectorVal,
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{},
				},
			},
			expect: nil,
		},
	} {
		if test.mutatePodFn != nil {
			test.mutatePodFn(test.pod)
		}
		executor := NewFairExecutor()
		actual, err := executor.OnPodCreating(test.selector, test.pod)
		if err != nil {
			t.Fatalf("[%s] [tolerations] executor on pod creating failed, err: %v", desc, err)
		}
		if len(actual) != len(test.expect) {
			t.Fatalf("[%s] [tolerations] executor on pod creating failed, actual: %d, expect %d", desc, len(actual), len(test.expect))
		}
		if len(actual) != 0 || len(test.expect) != 0 {
			for i := range actual {
				if actual[i].Op != test.expect[i].Op {
					t.Fatalf("[%s] [tolerations] executor on pod creating failed, actual: %s, expect %s", desc, actual[i].Op, test.expect[i].Op)
				}
				if actual[i].Path != test.expect[i].Path {
					t.Fatalf("[%s] [tolerations] executor on pod creating failed, actual: %s, expect %s", desc, actual[i].Path, test.expect[i].Path)
				}
				actualTolerations, ok := actual[i].Value.([]v1.Toleration)
				if !ok {
					t.Fatalf("[%s] [tolerations] executor on pod creating failed, actual: %v", desc, actual[i].Value)
				}
				expectTolerations, ok := test.expect[i].Value.([]v1.Toleration)
				if !ok {
					t.Fatalf("[%s] [tolerations] executor on pod creating failed, expect: %v", desc, test.expect[i].Value)
				}
				if len(actualTolerations) != len(expectTolerations) {
					t.Fatalf("[%s] [tolerations count] executor on pod creating failed, actual: %d, expect: %d", desc, len(actualTolerations), len(expectTolerations))
				}
				for j := range actualTolerations {
					if actualTolerations[j].Key != expectTolerations[j].Key {
						t.Fatalf("[%s] [tolerations key] executor on pod creating failed, actual: %s, expect: %s", desc, actualTolerations[j].Key, expectTolerations[j].Key)
					}
					if actualTolerations[j].Value != expectTolerations[j].Value {
						t.Fatalf("[%s] [tolerations value] executor on pod creating failed, actual: %s, expect: %s", desc, actualTolerations[j].Value, expectTolerations[j].Value)
					}
					if actualTolerations[j].Effect != expectTolerations[j].Effect {
						t.Fatalf("[%s] [tolerations effect] executor on pod creating failed, actual: %s, expect: %s", desc, actualTolerations[j].Effect, expectTolerations[j].Effect)
					}
					if actualTolerations[j].Operator != expectTolerations[j].Operator {
						t.Fatalf("[%s] [tolerations operator] executor on pod creating failed, actual: %s, expect: %s", desc, actualTolerations[j].Operator, expectTolerations[j].Operator)
					}
				}
			}
		}
	}
}

func TestFairOnPodCreatingWithAnnotationsAndLabels(t *testing.T) {
	for desc, test := range map[string]struct {
		pod         *v1.Pod
		selector    *eciv1.Selector
		mutatePodFn func(*v1.Pod)
		expect      []PatchInfo
	}{
		"test not exist selector annotations": {
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{},
				},
			},
			expect: nil,
		},
		"test exist selector annotations": {
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{
						Annotations: map[string]string{
							"foo": "boo",
						},
					},
				},
			},
			expect: []PatchInfo{
				{
					Op:    "add",
					Path:  "/metadata/annotations",
					Value: map[string]string{"foo": "boo"},
				},
			},
		},
		"test not exist selector labels": {
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{},
				},
			},
			expect: nil,
		},
		"test exist selector labels": {
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{
						Labels: map[string]string{
							"foo": "boo",
						},
					},
				},
			},
			expect: []PatchInfo{
				{
					Op:    "add",
					Path:  "/metadata/labels",
					Value: map[string]string{"foo": "boo"},
				},
			},
		},
	} {
		if test.mutatePodFn != nil {
			test.mutatePodFn(test.pod)
		}
		executor := NewFairExecutor()
		actual, err := executor.OnPodCreating(test.selector, test.pod)
		if err != nil {
			t.Fatalf("[%s] executor on pod creating failed, err: %v", desc, err)
		}
		if len(actual) != len(test.expect) {
			t.Fatalf("[%s] executor on pod creating failed, actual: %v, expect %v", desc, actual, test.expect)
		}
		if len(actual) != 0 || len(test.expect) != 0 {
			for i := range actual {
				if actual[i].Op != test.expect[i].Op {
					t.Fatalf("[%s] executor on pod creating failed, actual: %s, expect %s", desc, actual[i].Op, test.expect[i].Op)
				}
				if actual[i].Path != test.expect[i].Path {
					t.Fatalf("[%s] executor on pod creating failed, actual: %s, expect %s", desc, actual[i].Path, test.expect[i].Path)
				}
				actualAnnotations, ok := actual[i].Value.(map[string]string)
				if !ok {
					t.Fatalf("[%s] executor on pod creating failed, actual: %v", desc, actual[i].Value)
				}
				expectAnnotations, ok := test.expect[i].Value.(map[string]string)
				if !ok {
					t.Fatalf("[%s] executor on pod creating failed, expect: %v", desc, test.expect[i].Value)
				}
				if len(actualAnnotations) != len(expectAnnotations) {
					t.Fatalf("[%s] executor on pod creating failed, actual: %d, expect: %d", desc, len(actualAnnotations), len(expectAnnotations))
				}
				for k, v := range actualAnnotations {
					if v1, ok := expectAnnotations[k]; !ok || v != v1 {
						t.Fatalf("[%s] executor on pod creating failed, actual: %s, expect: %s", desc, actualAnnotations, expectAnnotations)
					}
				}
			}
		}
	}
}

func TestFairOnPodUnscheduled(t *testing.T) {
	for desc, test := range map[string]struct {
		pod         *v1.Pod
		selector    *eciv1.Selector
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{},
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
			selector: &eciv1.Selector{
				Spec: eciv1.SelectorSpec{
					Effect: &eciv1.SideEffect{
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
		executor := NewFairExecutor()
		actual, err := executor.OnPodUnscheduled(test.selector, test.pod)
		if err != nil && err != test.expectErr {
			t.Fatalf("[%s] executor on pod unscheduled failed, err: %v", desc, err)
		}
		if !reflect.DeepEqual(actual, test.expect) {
			t.Fatalf("[%s] executor on pod unscheduled failed, actual: %v, expect %v", desc, actual, test.expect)
		}
	}
}
