package main

import (

	//"crypto"
	//"crypto/ecdsa"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	v1 "k8s.io/api/admission/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	//"k8s.io/client-go/kubernetes"
	//kubernetes "github.com/rancher/rancher/pkg/generated/clientset/versioned/versioned"
	"github.com/rancher/lasso/pkg/client"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	cattlev1 "github.com/rancher/rancher/pkg/apis/ui.cattle.io/v1"
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
	client *client.Client
	//kubeClient rest.Interface
}

func (nls *NavlinksServerHandler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (nls *NavlinksServerHandler) serve(w http.ResponseWriter, r *http.Request) {

	//kubeClientSet kubernetes.Interface

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

	nav := createNavlinks(ns, "prometheus-operated", 9090)

	//_, err := nls.kubeClient.Post().Namespace(ns).Resource("navlink").Body(&nav).DoRaw(context.Background())
	err := nls.client.Create(context.TODO(), ns, &nav, &cattlev1.NavLink{}, metav1.CreateOptions{})

	//glog.Errorf("restresult %s", string(restresult))
	//UiV1().NavLinks(ns).Create(context.Context, &nav, metav1.CreateOptions{})
	//cattlev1.NavLinks(ns).Create(context.Context, nav, metav1.CreateOptions{})
	//   BatchV1().Jobs(newRds.Namespace).Create(context.Context, nav, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			glog.Error("navlinks already exists for ", ns)
			return
		}
		glog.Errorf("error creating navlinks: %v", err)
		return
	}

	createNavlinks(ns, "alertmanager-operated", 9093)
	createNavlinks(ns, "prometheus-monitoring-grafana", 80)
	glog.Error("navlinks create done", nav)

	resp, err := json.Marshal(admissionResponse(200, true, "Success", "Navlinks create", &arRequest))
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

func createNavlinks(namespace string, service string, port int) cattlev1.NavLink {
	return cattlev1.NavLink{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring-" + namespace,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "monitoring.coreos.com/v1",
					Kind:               "Prometheus",
					Controller:         &owner,
					BlockOwnerDeletion: &owner,
				},
			},
		},
		Spec: cattlev1.NavLinkSpec{
			Target: "_blank",
			Group:  "monitoring-" + namespace,
			ToService: &cattlev1.NavLinkTargetService{
				Namespace: namespace,
				Name:      service,
				Scheme:    "http",
				Port:      &intstr.IntOrString{IntVal: int32(port)},
				Path:      "",
			},
			//Icon: prometheus,
		},
	}
}

// Template for AdmissionReview
func admissionResponse(admissionCode int32, admissionPermissions bool, admissionStatus string, admissionMessage string, ar *v1.AdmissionReview) v1.AdmissionReview {
	return v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       admissionKind,
			APIVersion: admissionApi,
		},
		Response: &v1.AdmissionResponse{
			Allowed: admissionPermissions,
			UID:     ar.Request.UID,
			Result: &metav1.Status{
				Status:  admissionStatus,
				Message: admissionMessage,
				Code:    admissionCode,
			},
		},
	}
}
