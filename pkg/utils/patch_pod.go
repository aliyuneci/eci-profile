package utils

import (
	"context"
	"encoding/json"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type PatchOption struct {
	Metadata struct {
		Annotations map[string]string `json:"annotations,omitempty"`
		Labels      map[string]string `json:"labels,omitempty"`
	} `json:"metadata"`
	Spec struct {
		Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	} `json:"spec"`
}

func NewPatchOption() *PatchOption {
	return &PatchOption{}
}

func (o *PatchOption) WithAnnotations(annotations map[string]string) *PatchOption {
	if len(annotations) > 0 {
		o.Metadata.Annotations = annotations
	}
	return o
}

func (o *PatchOption) WithLabels(labels map[string]string) *PatchOption {
	if len(labels) > 0 {
		o.Metadata.Labels = labels
	}
	return o
}

func (o *PatchOption) WithTolerations(tolerations []v1.Toleration) *PatchOption {
	o.Spec.Tolerations = tolerations
	return o
}

func PatchPod(ctx context.Context, k8sClient *kubernetes.Clientset, namespace, name string, option PatchOption) (*v1.Pod, error) {
	payload, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}

	return k8sClient.CoreV1().Pods(namespace).Patch(ctx, name, types.MergePatchType, payload, metav1.PatchOptions{})
}
