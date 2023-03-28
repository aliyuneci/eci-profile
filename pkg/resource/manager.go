package resource

import (
	"time"

	eciv1 "eci.io/eci-profile/pkg/apis/eci/v1"
	"eci.io/eci-profile/pkg/client/clientset/versioned"
	"eci.io/eci-profile/pkg/client/informers/externalversions"
	listereciv1 "eci.io/eci-profile/pkg/client/listers/eci/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type Manager struct {
	coreV1InformerFactory  informers.SharedInformerFactory
	profileInformerFactory externalversions.SharedInformerFactory
	podInformer            cache.SharedIndexInformer
	nodeInformer           cache.SharedIndexInformer
	nsInformer             cache.SharedIndexInformer
	selectorInformer       cache.SharedIndexInformer
	rqInformer             cache.SharedIndexInformer
	podLister              listercorev1.PodLister
	nodeLister             listercorev1.NodeLister
	nsLister               listercorev1.NamespaceLister
	selectorLister         listereciv1.SelectorLister
	rqLister               listercorev1.ResourceQuotaLister
}

func NewManager(k8sClient *kubernetes.Clientset, profileClient *versioned.Clientset) *Manager {
	coreV1InformerFactory := informers.NewSharedInformerFactory(k8sClient, 30*time.Second)
	profileInformerFactory := externalversions.NewSharedInformerFactory(profileClient, 30*time.Second)
	return &Manager{
		coreV1InformerFactory:  coreV1InformerFactory,
		profileInformerFactory: profileInformerFactory,
		podInformer:            coreV1InformerFactory.Core().V1().Pods().Informer(),
		nodeInformer:           coreV1InformerFactory.Core().V1().Nodes().Informer(),
		nsInformer:             coreV1InformerFactory.Core().V1().Namespaces().Informer(),
		rqInformer:             coreV1InformerFactory.Core().V1().ResourceQuotas().Informer(),
		podLister:              coreV1InformerFactory.Core().V1().Pods().Lister(),
		nodeLister:             coreV1InformerFactory.Core().V1().Nodes().Lister(),
		nsLister:               coreV1InformerFactory.Core().V1().Namespaces().Lister(),
		rqLister:               coreV1InformerFactory.Core().V1().ResourceQuotas().Lister(),
		selectorInformer:       profileInformerFactory.Eci().V1().Selectors().Informer(),
		selectorLister:         profileInformerFactory.Eci().V1().Selectors().Lister(),
	}
}

func (m *Manager) Run(stopChan <-chan struct{}) {
	go m.coreV1InformerFactory.Start(stopChan)
	go m.profileInformerFactory.Start(stopChan)
}

func (m *Manager) HasSynced() bool {
	return m.podInformer.HasSynced() &&
		m.nodeInformer.HasSynced() &&
		m.nsInformer.HasSynced() &&
		m.rqInformer.HasSynced() &&
		m.selectorInformer.HasSynced()
}

func (m *Manager) AddPodEventHandler(handler cache.ResourceEventHandler) {
	m.podInformer.AddEventHandler(handler)
}

func (m *Manager) AddNodeEventHandler(handler cache.ResourceEventHandler) {
	m.nodeInformer.AddEventHandler(handler)
}

func (m *Manager) AddSelectorEventHandler(handler cache.ResourceEventHandler) {
	m.selectorInformer.AddEventHandler(handler)
}

func (m *Manager) AddNamespaceEventHandler(handler cache.ResourceEventHandler) {
	m.nsInformer.AddEventHandler(handler)
}

func (m *Manager) AddResourceQuotaEventHandler(handler cache.ResourceEventHandler) {
	m.rqInformer.AddEventHandler(handler)
}

func (m *Manager) ListNodes() ([]*v1.Node, error) {
	return m.nodeLister.List(labels.Everything())
}

func (m *Manager) GetNode(name string) (*v1.Node, error) {
	return m.nodeLister.Get(name)
}

func (m *Manager) ListPods(namespace string) ([]*v1.Pod, error) {
	if namespace != "" {
		return m.podLister.Pods(namespace).List(labels.Everything())
	}
	return m.podLister.List(labels.Everything())
}

func (m *Manager) GetPod(namespace, name string) (*v1.Pod, error) {
	return m.podLister.Pods(namespace).Get(name)
}

func (m *Manager) ListSelectors() ([]*eciv1.Selector, error) {
	return m.selectorLister.List(labels.Everything())
}

func (m *Manager) GetSelector(name string) (*eciv1.Selector, error) {
	return m.selectorLister.Get(name)
}

func (m *Manager) ListNamespaces() ([]*v1.Namespace, error) {
	return m.nsLister.List(labels.Everything())
}

func (m *Manager) GetNamespace(name string) (*v1.Namespace, error) {
	return m.nsLister.Get(name)
}

func (m *Manager) ListResourceQuotas(namespace string) ([]*v1.ResourceQuota, error) {
	if namespace != "" {
		return m.rqLister.ResourceQuotas(namespace).List(labels.Everything())
	}
	return m.rqLister.List(labels.Everything())
}

func (m *Manager) GetResourceQuota(namespace, name string) (*v1.ResourceQuota, error) {
	return m.rqLister.ResourceQuotas(namespace).Get(name)
}
