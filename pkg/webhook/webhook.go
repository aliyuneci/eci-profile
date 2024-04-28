package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (s *Server) registerMutatingWebhook(ctx context.Context) error {
	if s.isSupportAdmissionV1 {
		return s.registerMutatingWebhookV1(ctx)
	}
	return s.registerMutatingWebhookV1beta1(ctx)
}

func (s *Server) registerMutatingWebhookV1(ctx context.Context) error {
	client := s.k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations()
	if err := s.k8sClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(ctx, s.mutatingName, metav1.DeleteOptions{}); err != nil && !api_errors.IsNotFound(err) {
		klog.Warningf("[v1] delete %q V1Beta1 MutatingWebhookConfiguration failed: %s", s.mutatingName, err)
	}
	webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.mutatingName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{s.createV1MutatingWebhook(nil, nil)},
	}
	if _, err := client.Get(ctx, s.mutatingName, metav1.GetOptions{}); err != nil {
		if !api_errors.IsNotFound(err) {
			klog.Warningf("[v1] get %q MutatingAdmission failed: %s", s.mutatingName, err)
			return errors.Wrapf(err, "get '%s' mutating admission failed", s.mutatingName)
		}
		klog.Infof("[v1] create %q MutatingWebhookConfiguration ......", s.mutatingName)
		if _, err := client.Create(ctx, webhookConfig, metav1.CreateOptions{}); err != nil {
			klog.Errorf("[v1] create %q MutatingWebhookConfiguration failed: %s", s.mutatingName, err)
			return err
		} else {
			klog.Infof("[v1] create %q MutatingWebhookConfiguration.", s.mutatingName)
			return nil
		}
	}
	valueByte, _ := json.Marshal(webhookConfig.Webhooks)
	patchData := fmt.Sprintf(`[{"op":"replace","path":"/webhooks","value": %s}]`, string(valueByte))
	if _, err := client.Patch(ctx, s.mutatingName, types.JSONPatchType, []byte(patchData), metav1.PatchOptions{}); err != nil {
		klog.Errorf("Error patching MutatingWebhookConfiguration %q: %s", s.mutatingName, err)
		return fmt.Errorf("error patching MutatingWebhookConfiguration %q: %s", s.mutatingName, err)
	}
	klog.Infof("Patched MutatingWebhookConfiguration %q ...", s.mutatingName)
	return nil
}

func (s *Server) registerMutatingWebhookV1beta1(ctx context.Context) error {
	client := s.k8sClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
	webhookConfig := &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.mutatingName,
		},
		Webhooks: []admissionregistrationv1beta1.MutatingWebhook{s.createV1beta1MutatingWebhook(nil, nil)},
	}
	if _, err := client.Get(ctx, s.mutatingName, metav1.GetOptions{}); err != nil {
		if !api_errors.IsNotFound(err) {
			klog.Warningf("[v1] get %q MutatingAdmission failed: %s", s.mutatingName, err)
			return errors.Wrapf(err, "get '%s' mutating admission failed", s.mutatingName)
		}
		klog.Infof("[v1] create %q MutatingWebhookConfiguration ......", s.mutatingName)
		if _, err := client.Create(ctx, webhookConfig, metav1.CreateOptions{}); err != nil {
			klog.Errorf("[v1] create %q MutatingWebhookConfiguration failed: %s", s.mutatingName, err)
			return err
		}
	}
	return nil
}

func (s *Server) createV1MutatingWebhook(nsSelector, objectSelector *metav1.LabelSelector) admissionregistrationv1.MutatingWebhook {
	var (
		defaultSideEffectClass               = admissionregistrationv1.SideEffectClassNoneOnDryRun
		defaultFailurePolicy                 = admissionregistrationv1.Ignore
		defaultMatchPolicy                   = admissionregistrationv1.Equivalent
		defaultTimeoutSeconds          int32 = 5
		defaultAdmissionReviewVersions       = []string{"v1", "v1beta1"}
		defaultReinvocationPolicy            = admissionregistrationv1.NeverReinvocationPolicy
	)

	clientConfig := admissionregistrationv1.WebhookClientConfig{
		CABundle: s.certIssuer.GetCAData(),
		Service: &admissionregistrationv1.ServiceReference{
			Namespace: "kube-system",
			Name:      s.mutatingName,
			Path:      &s.serverPath,
			Port:      &s.serverPort,
		},
	}

	ruleOperation := []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"pods", "pods/binding"},
				Scope: func() *admissionregistrationv1.ScopeType {
					tmp := admissionregistrationv1.AllScopes
					return &tmp
				}(),
			},
		},
	}

	return admissionregistrationv1.MutatingWebhook{
		Name:                    "eci-profile.eci.aliyun.com",
		ClientConfig:            clientConfig,
		Rules:                   ruleOperation,
		FailurePolicy:           &defaultFailurePolicy,
		MatchPolicy:             &defaultMatchPolicy,
		NamespaceSelector:       nsSelector,
		ObjectSelector:          objectSelector,
		SideEffects:             &defaultSideEffectClass,
		TimeoutSeconds:          &defaultTimeoutSeconds,
		AdmissionReviewVersions: defaultAdmissionReviewVersions,
		ReinvocationPolicy:      &defaultReinvocationPolicy,
	}
}

func (s *Server) createV1beta1MutatingWebhook(nsSelector, objectSelector *metav1.LabelSelector) admissionregistrationv1beta1.MutatingWebhook {
	var (
		defaultSideEffectClass               = admissionregistrationv1beta1.SideEffectClassUnknown
		defaultFailurePolicy                 = admissionregistrationv1beta1.Ignore
		defaultMatchPolicy                   = admissionregistrationv1beta1.Equivalent
		defaultTimeoutSeconds          int32 = 5
		defaultAdmissionReviewVersions       = []string{"v1beta1"}
		defaultReinvocationPolicy            = admissionregistrationv1beta1.NeverReinvocationPolicy
	)

	clientConfig := admissionregistrationv1beta1.WebhookClientConfig{
		CABundle: s.certIssuer.GetCAData(),
		Service: &admissionregistrationv1beta1.ServiceReference{
			Namespace: "kube-system",
			Name:      s.mutatingName,
			Path:      &s.serverPath,
			Port:      &s.serverPort,
		},
	}

	ruleOperation := []admissionregistrationv1beta1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create},
			Rule: admissionregistrationv1beta1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"pods", "pods/binding"},
				Scope: func() *admissionregistrationv1beta1.ScopeType {
					tmp := admissionregistrationv1beta1.AllScopes
					return &tmp
				}(),
			},
		},
	}

	return admissionregistrationv1beta1.MutatingWebhook{
		Name:                    "autoscaler.eci.aliyun.com",
		ClientConfig:            clientConfig,
		Rules:                   ruleOperation,
		FailurePolicy:           &defaultFailurePolicy,
		MatchPolicy:             &defaultMatchPolicy,
		NamespaceSelector:       nsSelector,
		ObjectSelector:          objectSelector,
		SideEffects:             &defaultSideEffectClass,
		TimeoutSeconds:          &defaultTimeoutSeconds,
		AdmissionReviewVersions: defaultAdmissionReviewVersions,
		ReinvocationPolicy:      &defaultReinvocationPolicy,
	}
}
