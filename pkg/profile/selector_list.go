package profile

import eciv1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"

type SelectorList []eciv1beta1.Selector

func (sl SelectorList) Less(i, j int) bool {
	iPriority := int32(0)
	jPriority := int32(0)
	if sl[i].Spec.Priority != nil {
		iPriority = *sl[i].Spec.Priority
	}
	if sl[j].Spec.Priority != nil {
		jPriority = *sl[j].Spec.Priority
	}
	return iPriority > jPriority
}
func (sl SelectorList) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}
func (sl SelectorList) Len() int {
	return len(sl)
}
