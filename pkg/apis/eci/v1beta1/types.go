package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Selector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SelectorSpec `json:"spec"`
}

type SelectorSpec struct {
	NamespaceLabels *metav1.LabelSelector `json:"namespaceLabels,omitempty"`
	ObjectLabels    *metav1.LabelSelector `json:"objectLabels,omitempty"`
	Effect          *SideEffect           `json:"effect,omitempty"`
	Policy          *PolicySource         `json:"policy,omitempty"`
	Priority        *int32                `json:"priority,omitempty"`
}

type FairPolicySource struct{}

type NormalNodeOnlyPolicySource struct{}

type NormalNodePreferPolicySource struct {
	CPURatio    *float64 `json:"cpuRatio,omitempty"`
	MemoryRatio *float64 `json:"memoryRatio,omitempty"`
}

type VirtualNodeOnlyPolicySource struct{}

type NamespaceResourceLimitPolicySource struct {
	Namespace string          `json:"namespace"`
	Limits    v1.ResourceList `json:"limits"`
}

type PolicySource struct {
	Fair                   *FairPolicySource                   `json:"fair,omitempty"`
	NormalNodeOnly         *NormalNodeOnlyPolicySource         `json:"normalNodeOnly,omitempty"`
	NormalNodePrefer       *NormalNodePreferPolicySource       `json:"normalNodePrefer,omitempty"`
	VirtualNodeOnly        *VirtualNodeOnlyPolicySource        `json:"virtualNodeOnly,omitempty"`
	NamespaceResourceLimit *NamespaceResourceLimitPolicySource `json:"namespaceResourceLimit,omitempty"`
}

type SideEffect struct {
	Annotations map[string]string `json:"annotations,omitempty"` // 需要追加的annotation
	Labels      map[string]string `json:"labels,omitempty"`      // 需要追加的label
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SelectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Selector `json:"items"`
}
