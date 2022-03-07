package ingress

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.mgmt.arms-dev.net/go-common/gkeclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Service interface {
	UseAnnotations() (bool, error)
}

type service struct {
	log   *logrus.Entry
	k8scs *kubernetes.Clientset
}

func New(log *logrus.Entry) (Service, error) {
	cs, err := gkeclient.New()
	if err != nil {
		return nil, err
	}
	return &service{
		log:   log,
		k8scs: cs,
	}, nil
}

func (s *service) UseAnnotations() (bool, error) {

	version, err := s.getControllerVersion()
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(version, "0."), nil
}

func (s *service) getControllerVersion() (string, error) {

	const (
		namespace      = "nginx-ingress"
		deployment     = "nginx-ingress-internal-controller"
		label          = "app.kubernetes.io/version"
		defaultVersion = "0.24.1"
	)

	dep, err := s.k8scs.AppsV1().Deployments(namespace).Get(context.Background(), deployment, v1.GetOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return "", err
	}

	if dep == nil || dep.Labels == nil {
		return defaultVersion, nil
	}

	if dep.Labels[label] != "" {
		return strings.TrimSpace(dep.Labels[label]), nil
	}

	return defaultVersion, nil
}
