/*
NodeSelector-enforcer is a mutating admission webhook (https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)
meant to be deployed as a Google Function (https://cloud.google.com/functions/).

It will respond to AdmissionReview requests from kubernetes clusters configured using this MAW and
send pods to nodes with matching labels (https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector).
*/

package nodeselector

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// stringInSlice is like in python: if 'string' in list
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// getEnv will get the value of the environment variable, in case of missing or empty, use the provided fallback value
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// addOrReplaceNodeSelector returns a patch, which adds or replaces the node selector field in the pod definition
func addOrReplaceNodeSelector(pod *corev1.Pod, namespace string, basePath string) (patch []patchOperation) {
	var systemNamespaces = strings.Split(getEnv("SYSTEM_NAMESPACES", "istio-system,kube-system,ingress"), ",")
	var patchOp string
	var purpose string

	if len(pod.Spec.NodeSelector) > 0 {
		patchOp = "replace"
	} else {
		patchOp = "add"
	}

	switch {
	case stringInSlice(namespace, systemNamespaces):
		purpose = "system-processes"
	case strings.HasPrefix(namespace, "prod-"):
		purpose = "production"
	default:
		purpose = "non-production"
	}

	patch = append(patch, patchOperation{
		Op:    patchOp,
		Path:  basePath,
		Value: map[string]string{"purpose": purpose},
	})
	return patch
}

// mutate adds the nodeSelector field to the pod spec
func mutate(ar *admissionv1beta1.AdmissionReview) *admissionv1beta1.AdmissionResponse {
	var pod corev1.Pod

	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	req := ar.Request
	log.Printf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo.Username)

	patchBytes, err := json.Marshal(addOrReplaceNodeSelector(&pod, req.Namespace, "/spec/nodeSelector"))
	if err != nil {
		return &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Printf("AdmissionResponse: patch=%v", string(patchBytes))
	return &admissionv1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1beta1.PatchType {
			pt := admissionv1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// EntryPoint reads body, parses the AdmissionReview and returns AdmissionResponse
func EntryPoint(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	if len(data) > 0 {
		var admissionRequest admissionv1beta1.AdmissionReview
		var admissionResponse *admissionv1beta1.AdmissionResponse
		if err := json.Unmarshal(data, &admissionRequest); err != nil {
			log.Printf("Cannot decode body: %v", err)
			admissionResponse = &admissionv1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		} else {
			log.Printf("Patch request for %v", admissionRequest.Request.UID)
			admissionResponse = mutate(&admissionRequest)
		}
		admissionReview := admissionv1beta1.AdmissionReview{}
		if admissionResponse != nil {
			admissionReview.Response = admissionResponse
			if admissionRequest.Request != nil {
				admissionReview.Response.UID = admissionRequest.Request.UID
			}
		}

		resp, err := json.Marshal(admissionReview)
		if err != nil {
			log.Printf("Cannot encode response: %v", err)
			http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		}
		log.Printf("Patch created for %v", admissionRequest.Request.UID)
		if _, err := w.Write(resp); err != nil {
			log.Printf("Cannot write response: %v", err)
			http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		}
		return
	}
	io.WriteString(w, getEnv("VERSION", "no version was specified"))
}
