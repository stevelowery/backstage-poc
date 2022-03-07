package main

import (
	"encoding/json"
	"net/http"

	"gitlab.mgmt.arms-dev.net/go-common/healthcheck"

	"github.com/sirupsen/logrus"
	"gitlab.mgmt.arms-dev.net/go-common/logger"
	"gitlab.mgmt.arms-dev.net/infrastructure/ingress-class-webhook/internal/controller"

	k8s "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretsPath = "/var/run/secrets/certs"
)

type service struct {
	log *logrus.Entry
}

func main() {
	log := logger.NewEntry().WithField("context", "ingress-class-webhook")
	logger.InfoAlways(log, "Listening for requests...")

	svc := &service{
		log: log,
	}

	mux := http.NewServeMux()
	hc := healthcheck.NewServer(healthcheck.Opts{
		AppLogEntry: log,
	})

	mux.Handle("/liveness", hc.Liveness())
	mux.Handle("/readiness", hc.Readiness())
	mux.Handle("/", http.HandlerFunc(svc.handle))

	if err := http.ListenAndServeTLS(":443", secretsPath+"/tls.crt", secretsPath+"/tls.key", mux); err != nil {
		log.Fatal(err)
	}
}

func (s *service) handle(rw http.ResponseWriter, r *http.Request) {

	input := &k8s.AdmissionReview{}

	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		s.log.Errorf("Error reading body: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}

	ctrl, err := controller.New(s.log)
	if err != nil {
		s.log.Errorf("Error creating controller: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}

	response := ctrl.Process(input.Request)
	output := &k8s.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			APIVersion: input.APIVersion,
			Kind:       input.Kind,
		},
		Response: response,
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(output)

	if err != nil {
		s.log.Errorf("Error encoding response: %v", err)
	}

	s.log.Debugf("Request %s for resource %s/%s was accepted %v", input.Request.UID, input.Request.Namespace, input.Request.Name, response.Allowed)
}
