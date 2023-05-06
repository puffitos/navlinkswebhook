package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/client-go/rest"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/admission/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionApi  = "admission.k8s.io/v1"
	admissionKind = "AdmissionReview"
)

var (
	owner = bool(true)
)

// NavlinksServerHandler listen to admission requests and serve responses
type NavlinksServerHandler struct {
}

func (nls *NavlinksServerHandler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (nls *NavlinksServerHandler) serve(w http.ResponseWriter, r *http.Request) {

	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// Url path of metrics
	if r.URL.Path == "/metrics" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Url path of admission
	if r.URL.Path != "/validate" {
		glog.Error("no validate")
		http.Error(w, "no validate", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// count each request for prometheus metric
	opsProcessed.Inc()
	arRequest := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &arRequest); err != nil {
		glog.Error("incorrect body")
		http.Error(w, "incorrect body", http.StatusBadRequest)
		return
	}

	raw := arRequest.Request.Object.Raw
	prom := monitoringv1.Prometheus{}
	if err := json.Unmarshal(raw, &prom); err != nil {
		glog.Error("error deserializing pod")
		return
	}

	ns := prom.Namespace
	glog.Error("prom namespace", ns)

	if len(ns) == 0 {
		glog.Errorf("No namespace found %s/%s", prom.Name, prom.Namespace)
		resp, err := json.Marshal(admissionResponse(200, true, "Success", "Navlinks create skipped", &arRequest))
		if err != nil {
			glog.Errorf("Can't encode response: %v", err)
			http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		}
		if _, err := w.Write(resp); err != nil {
			glog.Errorf("Can't write response: %v", err)
			http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Error("cant get incluster config: ", err)
		return
	}
	// creates the clientset
	clientset := NewForConfigOrDie(config)
	if err != nil {
		glog.Error("cant setup clientset: ", ns)
		return
	}

	// check if navlink resource is available on api server
	_, err = clientset.Navlinks().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		glog.Error("navlinks resource not available: ", err)
		return
	}

	// create navlink resource prometheus-operated
	navPrometheus := createNavlinks(ns, "prometheus-operated", "9090", string(arRequest.Request.UID))
	_, err = clientset.Navlinks().Create(context.TODO(), &navPrometheus, metav1.CreateOptions{})

	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			glog.Error("navlinks prometheus already exists for ", ns)
			nls.response(true, "Navlink prometheus already exists, skipped", w, &arRequest)
		}
		glog.Errorf("error creating navlinks: %v", err)
		return
	}

	glog.Error("navlinks created: ", navPrometheus.Name)

	// create navlink resource alertmanager-operated
	navAlertManager := createNavlinks(ns, "alertmanager-operated", "9093", string(arRequest.Request.UID))
	_, err = clientset.Navlinks().Create(context.TODO(), &navAlertManager, metav1.CreateOptions{})

	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			glog.Error("navlinks alertmanager already exists for ", ns)
			nls.response(true, "Navlink alertmanager already exists, skipped", w, &arRequest)
		}
		glog.Errorf("error creating navlinks: %v", err)
		return
	}
	glog.Error("navlinks created: ", navAlertManager.Name)

	// create navlink resource prometheus-monitoring-grafana
	navGrafana := createNavlinks(ns, "prometheus-monitoring-grafana", "80", string(arRequest.Request.UID))
	_, err = clientset.Navlinks().Create(context.TODO(), &navGrafana, metav1.CreateOptions{})

	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			glog.Error("navlinks grafana already exists for ", ns)
			nls.response(true, "Navlink grafana already exists, skipped", w, &arRequest)
		}
		glog.Errorf("error creating navlinks: %v", err)
		return
	}
	glog.Error("navlinks create done", navGrafana.Name)

	resp, err := json.Marshal(admissionResponse(200, true, "Success", "Navlinks create", &arRequest))
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (nls *NavlinksServerHandler) response(allowed bool, message string, w http.ResponseWriter, arRequest *v1.AdmissionReview) {

	resp, err := json.Marshal(admissionResponse(200, allowed, "Success", message, arRequest))
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
