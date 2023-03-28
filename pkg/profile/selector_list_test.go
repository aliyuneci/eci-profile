package profile

import (
	"testing"

	"sort"

	v1 "eci.io/eci-profile/pkg/apis/eci/v1"
)

func TestSelectorList(t *testing.T) {
	int5 := int32(5)
	int4 := int32(4)
	int3 := int32(3)
	int2 := int32(2)
	int1 := int32(1)
	for desc, test := range map[string]struct {
		originSelectorList SelectorList
		expectSelectorList SelectorList
	}{
		"test selector list 1": {
			originSelectorList: SelectorList{
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int5}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int4}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int3}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int2}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int1}},
			},
			expectSelectorList: SelectorList{
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int5}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int4}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int3}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int2}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int1}},
			},
		},
		"test selector list 2": {
			originSelectorList: SelectorList{
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int1}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int3}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int4}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int2}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int5}},
			},
			expectSelectorList: SelectorList{
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int5}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int4}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int3}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int2}},
				v1.Selector{Spec: v1.SelectorSpec{Priority: &int1}},
			},
		},
	} {
		sort.Sort(test.originSelectorList)
		for i := range test.originSelectorList {
			if test.originSelectorList[i].Spec.Priority != test.expectSelectorList[i].Spec.Priority {
				t.Fatalf("%s test failed, origin: %v, expect: %v", desc, test.originSelectorList, test.expectSelectorList)
			}
		}
	}
}
