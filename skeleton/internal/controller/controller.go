package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.mgmt.arms-dev.net/infrastructure/ingress-class-webhook/internal/ingress"
	admission "k8s.io/api/admission/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressClassAnnotationKey          = "kubernetes.io/ingress.class"
	defaultIngressClassAnnotationValue = "nginx"
	defaultIngressClassName            = "nginx-abs"
	kong                               = "kong"
)

type Service interface {
	Process(req *admission.AdmissionRequest) *admission.AdmissionResponse
}

type service struct {
	log            *logrus.Entry
	ingressService ingress.Service
}

func New(log *logrus.Entry) (Service, error) {

	is, err := ingress.New(log)
	if err != nil {
		return nil, err
	}

	return &service{
		log:            log,
		ingressService: is,
	}, nil
}

func (s *service) Process(req *admission.AdmissionRequest) *admission.AdmissionResponse {

	const (
		ext     = "extensions/v1beta1"
		netbeta = "networking.k8s.io/v1beta1"
		netv1   = "networking.k8s.io/v1"
	)

	var value string
	var annotations map[string]string
	var ingressClassName *string

	useAnnotation, err := s.ingressService.UseAnnotations()
	if err != nil {
		return s.toResponse(req, useAnnotation, "", nil, nil, err)
	}

	version, err := s.getApiVersion(req)
	if err != nil {
		return s.toResponse(req, useAnnotation, "", nil, nil, err)
	}

	switch version {
	case ext:
		obj := &extensionsv1beta1.Ingress{}
		err = s.unmarshallTo(req, obj)
		if err != nil {
			return s.toResponse(req, useAnnotation, "", nil, nil, err)
		}
		annotations = obj.ObjectMeta.Annotations
		ingressClassName = obj.Spec.IngressClassName
	case netbeta:
		obj := &networkingv1beta1.Ingress{}
		err = s.unmarshallTo(req, obj)
		if err != nil {
			return s.toResponse(req, useAnnotation, "", nil, nil, err)
		}
		annotations = obj.ObjectMeta.Annotations
		ingressClassName = obj.Spec.IngressClassName
	case netv1:
		obj := &networkingv1.Ingress{}
		err = s.unmarshallTo(req, obj)
		if err != nil {
			return s.toResponse(req, useAnnotation, "", nil, nil, err)
		}
		annotations = obj.ObjectMeta.Annotations
		ingressClassName = obj.Spec.IngressClassName
	}

	value = s.getIngressClassName(annotations, ingressClassName, useAnnotation)
	return s.toResponse(req, useAnnotation, value, annotations, ingressClassName, nil)
}

func (s *service) getApiVersion(req *admission.AdmissionRequest) (string, error) {

	const apiVersion = "apiVersion"

	raw := map[string]interface{}{}
	err := json.Unmarshal(req.Object.Raw, &raw)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", raw[apiVersion]), nil
}

func (s *service) getIngressClassName(annotations map[string]string, ingressClassName *string, useAnnotations bool) string {
	var calculated string
	if ingressClassName != nil && *ingressClassName != "" {
		calculated = *ingressClassName
	} else if annotations != nil && annotations[ingressClassAnnotationKey] != "" {
		calculated = annotations[ingressClassAnnotationKey]
	}

	if useAnnotations && (calculated == defaultIngressClassName || calculated == "") {
		calculated = defaultIngressClassAnnotationValue
	} else if !useAnnotations && (calculated == defaultIngressClassAnnotationValue || calculated == "") {
		calculated = defaultIngressClassName
	}
	return calculated
}

func (s *service) unmarshallTo(req *admission.AdmissionRequest, obj interface{}) error {
	err := json.Unmarshal(req.Object.Raw, obj)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) toResponse(req *admission.AdmissionRequest, useAnnotation bool, value string, annotations map[string]string, ingressClassName *string, err error) *admission.AdmissionResponse {
	res := &admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: err == nil,
	}

	if err != nil {
		res.Result = &metav1.Status{
			Message: err.Error(),
		}
		s.log.Errorf("Error returned to client: %v", err)
		return res
	}

	// don't modify kong ingress resources
	if !strings.HasPrefix(value, kong) {
		pt := admission.PatchTypeJSONPatch
		res.PatchType = &pt
		res.Patch = toJSONPatch(useAnnotation, value, annotations, ingressClassName)
	}

	return res
}

func toJSONPatch(useAnnotation bool, value string, annotations map[string]string, ingressClassName *string) []byte {

	type patchOperation struct {
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value,omitempty"`
	}

	const (
		ingressClassAnnotationPath = "/metadata/annotations/kubernetes.io~1ingress.class"
		ingressClassNamePath       = "/spec/ingressClassName"
	)

	p := []*patchOperation{}

	if useAnnotation {
		p = append(p, &patchOperation{
			Op:    "replace",
			Path:  ingressClassAnnotationPath,
			Value: value,
		})
		if ingressClassName != nil {
			p = append(p, &patchOperation{
				Op:   "remove",
				Path: ingressClassNamePath,
			})
		}
	} else {
		p = append(p, &patchOperation{
			Op:    "replace",
			Path:  ingressClassNamePath,
			Value: value,
		})
		if annotations != nil && annotations[ingressClassAnnotationKey] != "" {
			p = append(p, &patchOperation{
				Op:   "remove",
				Path: ingressClassAnnotationPath,
			})
		}
	}

	patch, _ := json.Marshal(p)
	return patch
}
