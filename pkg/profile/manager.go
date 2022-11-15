package profile

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	eciv1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"
	"eci.io/eci-profile/pkg/client/clientset/versioned"
	"eci.io/eci-profile/pkg/policy"
	"eci.io/eci-profile/pkg/resource"
	"eci.io/eci-profile/pkg/utils"
	"eci.io/eci-profile/pkg/webhook"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Config struct {
	K8sClient     *kubernetes.Clientset
	ProfileClient *versioned.Clientset
	CACertPath    string
	CAKeyPath     string
}

type Manager struct {
	resourceManager *resource.Manager
	policyManager   *policy.Manager
	webhookServer   *webhook.Server
	k8sClient       *kubernetes.Clientset
}

func NewManager(config *Config) (*Manager, error) {
	resourceManager := resource.NewManager(config.K8sClient, config.ProfileClient)
	policyManager := policy.NewManager(resourceManager)
	manager := &Manager{
		resourceManager: resourceManager,
		policyManager:   policyManager,
		k8sClient:       config.K8sClient,
	}

	webhookConfig := &webhook.Config{
		K8sClient:     config.K8sClient,
		MutatePodFunc: manager.onPodCreating,
		CACertPath:    config.CACertPath,
		CAKeyPath:     config.CAKeyPath,
	}
	webhookServer, err := webhook.NewServer(webhookConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create webhook server")
	}
	manager.webhookServer = webhookServer
	manager.registerPodEventHandler()
	return manager, nil
}

func (m *Manager) Run(ctx context.Context) error {
	klog.Info("ready to start resource manager service")
	m.resourceManager.Run(ctx.Done())
	klog.Info("waiting for resource manager cache syncing")
	cache.WaitForCacheSync(ctx.Done(), m.resourceManager.HasSynced)
	klog.Info("resource manager cache has synced")
	return m.webhookServer.Run(ctx)
}

func (m *Manager) onPodCreating(pod *v1.Pod) ([]policy.PatchInfo, error) {
	selector, err := m.matchSelectorForPod(pod)
	if err != nil {
		return nil, errors.Wrap(err, "failed to match selector")
	}
	if selector == nil {
		klog.V(3).Infof("no selector matched for pod %s/%s, skip it", pod.Namespace, pod.Name)
		return nil, nil
	}
	klog.Infof("pod %s/%s(%s) matched the selector %s(%s)", pod.Namespace, pod.Name, pod.UID, selector.Name, selector.UID)
	return m.policyManager.OnPodCreating(selector, pod)
}

func (m *Manager) onPodUnscheduled(pod *v1.Pod) error {
	klog.V(3).Infof("pod %s/%s is unscheduled, recheck it", pod.Namespace, pod.Name)
	selector, err := m.matchSelectorForPod(pod)
	if err != nil {
		return errors.Wrap(err, "failed to match selector")
	}
	if selector == nil {
		klog.V(3).Infof("no selector matched for pod %s/%s", pod.Namespace, pod.Name)
		return nil
	}

	klog.Infof("pod %s/%s(%s) matched the selector %s(%s)", pod.Namespace, pod.Name, pod.UID, selector.Name, selector.UID)
	patchOptions, err := m.policyManager.OnPodUnscheduled(selector, pod)
	if err != nil {
		return errors.Wrap(err, "execute policy failed")
	}
	if patchOptions != nil {
		if _, err := utils.PatchPod(context.TODO(), m.k8sClient, pod.Namespace, pod.Name, *patchOptions); err != nil {
			klog.Errorf("failed to patch the pod %s/%s(%s): %q", pod.Namespace, pod.Name, pod.UID, err)
			return errors.Wrap(err, "failed to patch pod")
		}
		klog.Infof("the pod %s/%s is allowed to schedule to vnode (matched: %s)", pod.Namespace, pod.Name, selector.Name)
	}
	return nil
}

func (m *Manager) matchSelectorForPod(pod *v1.Pod) (*eciv1beta1.Selector, error) {
	allSelectors, err := m.resourceManager.ListSelectors()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list selectors")
	}
	var selectors []eciv1beta1.Selector
	for _, selector := range allSelectors {
		matched, err := m.matchPod(selector, pod)
		if err != nil {
			return nil, errors.Wrap(err, "match pod failed")
		}
		if matched {
			selectors = append(selectors, *selector)
		}
	}
	var selector *eciv1beta1.Selector
	if len(selectors) > 0 {
		sort.Sort(SelectorList(selectors))
		selector = &selectors[0]
	}
	return selector, nil
}

func (m *Manager) registerPodEventHandler() {
	m.resourceManager.AddSelectorEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			selector, ok := obj.(*eciv1beta1.Selector)
			if !ok {
				return
			}
			klog.Infof("add selector: %s(%s)", selector.Name, selector.UID)
			payload, _ := json.Marshal(selector)
			klog.V(5).Infof("selector payload: %s", payload)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if reflect.DeepEqual(oldObj, newObj) {
				return
			}
			selector, ok := newObj.(*eciv1beta1.Selector)
			if !ok {
				return
			}
			klog.Infof("update selector: %s(%s)", selector.Name, selector.UID)
			payload, _ := json.Marshal(selector)
			klog.V(5).Infof("selector payload: %s", payload)
		},
		DeleteFunc: func(obj interface{}) {
			selector, ok := obj.(*eciv1beta1.Selector)
			if !ok {
				return
			}
			klog.Infof("delete selector: %s(%s)", selector.Name, selector.UID)
		},
	})
	m.resourceManager.AddPodEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return
			}
			if isUnscheduledPod(pod) {
				if err := m.onPodUnscheduled(pod); err != nil {
					klog.Errorf("failed to execute unscheduled policy for pod %s/%s: %q", pod.Namespace, pod.Name, err)
					return
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}
			if isUnscheduledPod(pod) {
				if err := m.onPodUnscheduled(pod); err != nil {
					klog.Errorf("failed to execute unscheduled policy for pod %s/%s: %q", pod.Namespace, pod.Name, err)
					return
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
}

func (m *Manager) matchPod(selector *eciv1beta1.Selector, pod *v1.Pod) (bool, error) {
	if selector.Spec.NamespaceLabels != nil {
		selector, err := metav1.LabelSelectorAsSelector(selector.Spec.NamespaceLabels)
		if err != nil {
			return false, errors.Wrap(err, "failed to convert namespace selector to selector")
		}
		namespace, err := m.resourceManager.GetNamespace(pod.Namespace)
		if err != nil {
			return false, errors.Wrap(err, "failed to get namespace labels")
		}
		namespaceLabels := namespace.Labels
		if !selector.Matches(labels.Set(namespaceLabels)) {
			return false, nil
		}
	}
	if selector.Spec.ObjectLabels != nil {
		selector, err := metav1.LabelSelectorAsSelector(selector.Spec.ObjectLabels)
		if err != nil {
			return false, errors.Wrap(err, "failed to convert object selector to selector")
		}
		podLabels := pod.GetLabels()
		if !selector.Matches(labels.Set(podLabels)) {
			return false, nil
		}
	}
	return true, nil
}
