package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"eci.io/eci-profile/pkg/cert"
	"eci.io/eci-profile/pkg/policy"
	"github.com/pkg/errors"
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	mutatingName   = "eci-profile"
	supportVersion = "1.16.0"
)

// admitv1beta1Func handles a v1beta1 admission
type admitv1beta1Func func(admissionv1beta1.AdmissionReview) *admissionv1beta1.AdmissionResponse

// admitv1beta1Func handles a v1 admission
type admitv1Func func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

type MutatePodFunc func(pod *v1.Pod) ([]policy.PatchInfo, error)

type Config struct {
	K8sClient     *kubernetes.Clientset
	MutatePodFunc MutatePodFunc
	CACertPath    string
	CAKeyPath     string
}

type Server struct {
	isSupportAdmissionV1 bool
	k8sClient            *kubernetes.Clientset
	mutatingName         string
	serverPath           string
	serverPort           int32
	certIssuer           *cert.Issuer
	mutatePodFunc        MutatePodFunc
}

func NewServer(config *Config) (*Server, error) {
	isSupportAdmissionV1 := true
	serverVersion, err := config.K8sClient.DiscoveryClient.ServerVersion()
	if err != nil {
		return nil, errors.Wrap(err, "get cluster ServerVersion failed")
	}
	currentServerVersion, err := utilversion.ParseSemantic(serverVersion.String())
	if err != nil {
		return nil, errors.Wrap(err, "parse current server version failed")
	}
	supportServerVersion, err := utilversion.ParseSemantic(supportVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parse support server version failed")
	}
	if currentServerVersion.LessThan(supportServerVersion) {
		isSupportAdmissionV1 = false
	}
	klog.Infof("ServerVersion: %s, Major: %s, Minor: %s, SupportAdmissionV1: %v",
		serverVersion, serverVersion.Major, serverVersion.Minor, isSupportAdmissionV1)

	if config.CACertPath != "" && config.CAKeyPath != "" {
		caCertData, err := ioutil.ReadFile(config.CACertPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load CA cert file")
		}

		caKeyData, err := ioutil.ReadFile(config.CAKeyPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load CA key file")
		}

		klog.Infof("ready to create cert issuer with specified CA")
		caCert = caCertData
		caKey = caKeyData
	}
	certIssuer, err := cert.NewIssuer(caCert, caKey)
	if err != nil {
		klog.Errorf("failed to create cert issuer: %q", err)
		return nil, errors.Wrap(err, "failed to create cert issuer")
	}
	return &Server{
		isSupportAdmissionV1: isSupportAdmissionV1,
		k8sClient:            config.K8sClient,
		certIssuer:           certIssuer,
		mutatingName:         mutatingName,
		serverPath:           "/inject",
		serverPort:           443,
		mutatePodFunc:        config.MutatePodFunc,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	klog.Info("start to register mutating webhook")
	if err := s.registerMutatingWebhook(ctx); err != nil {
		klog.Errorf("failed to register mutating webhook: %q", err)
		return errors.Wrap(err, "failed to register webhook")
	}
	klog.Info("register mutating webhook successfully")

	hosts := []string{s.mutatingName,
		fmt.Sprintf("%s.kube-system", s.mutatingName),
		fmt.Sprintf("%s.kube-system.svc", s.mutatingName)}
	serverCertPemData, serverKeyPemData, err := s.certIssuer.IssueCSR(s.mutatingName, hosts)
	if err != nil {
		return errors.Wrap(err, "failed to issue server cert")
	}
	sCert, err := tls.X509KeyPair(serverCertPemData, serverKeyPemData)
	if err != nil {
		return errors.Wrap(err, "failed to load webhook server cert")
	}
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.serverPort),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{sCert},
		},
	}
	http.HandleFunc(s.serverPath, s.serveMutatingPod)
	http.HandleFunc("/healthz", s.healthCheckHandle)

	klog.Info("ready to start webhook http service")
	return server.ListenAndServeTLS("", "")
}

func (s *Server) serveMutatingPod(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(s.mutatePod))
}

func (s *Server) healthCheckHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// admitHandler is a handler, for both validators and mutators, that supports multiple admission review versions
type admitHandler struct {
	v1beta1 admitv1beta1Func
	v1      admitv1Func
}

func newDelegateToV1AdmitHandler(f admitv1Func) admitHandler {
	return admitHandler{
		v1beta1: delegateV1beta1AdmitToV1(f),
		v1:      f,
	}
}

func serve(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		msg := "Request could not empty body"
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("contentType=%s, expect application/json", contentType)
		klog.Errorf(msg)
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	klog.V(5).Infof("handling request: %s", body)

	deserializer := codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case admissionv1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*admissionv1beta1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1beta1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &admissionv1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	case admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*admissionv1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &admissionv1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(5).Infof("sending response: %v", responseObj)

	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func (s *Server) mutatePod(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v PatchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	if req.Resource != podResource {
		klog.Errorf("Resource=%s, expect Resource %s", req.Resource.String(), podResource.String())
		return nil
	}

	pod := &v1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(req.Object.Raw, nil, pod); err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	pod.Namespace = req.Namespace
	patchInfos, err := s.mutatePodFunc(pod)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	ret := &admissionv1.AdmissionResponse{
		Allowed: true,
	}
	if len(patchInfos) != 0 {
		data, _ := json.Marshal(patchInfos)
		klog.V(4).Infof("PatchData: %s", string(data))
		ret.PatchType = func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}()
		ret.Patch = data
	}
	return ret
}
