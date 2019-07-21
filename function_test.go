package nodeselector

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEntryPointGET(t *testing.T) {
	payload := strings.NewReader("")
	req := httptest.NewRequest("GET", "/", payload)

	rr := httptest.NewRecorder()
	EntryPoint(rr, req)

	out, err := ioutil.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	want := "no version was specified"
	if got := string(out); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEntryPointPOSTAdd(t *testing.T) {
	admissionReview := strings.NewReader(`{
		"kind": "AdmissionReview",
		"apiVersion": "admission.k8s.io/v1beta1",
		"request": {
			"uid": "test-no-nodeselector-present",
			"kind": {
				"group": "",
				"version": "v1",
				"kind": "Pod"
			},
			"resource": {
				"group": "",
				"version": "v1",
				"resource": "pods"
			},
			"namespace": "test",
			"name": "test",
			"operation": "CREATE",
			"userInfo": {
				"username": "go-test"
			},
			"object": {
				"metadata": {
					"generateName": "nginx-65899c769f-",
					"creationTimestamp": null,
					"labels": {
						"pod-template-hash": "2145573259",
						"run": "nginx"
					},
					"ownerReferences": [{
						"apiVersion": "extensions/v1beta1",
						"kind": "ReplicaSet",
						"name": "nginx-65899c769f",
						"uid": "56ff8242-cc6f-11e8-b891-42010a8e0fd6",
						"controller": true,
						"blockOwnerDeletion": true
					}]
				},
				"spec": {
					"volumes": [{
						"name": "default-token-wgwgs",
						"secret": {
							"secretName": "default-token-wgwgs"
						}
					}],
					"containers": [{
						"name": "nginx",
						"image": "nginx",
						"resources": {},
						"volumeMounts": [{
							"name": "default-token-wgwgs",
							"readOnly": true,
							"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
						}],
						"terminationMessagePath": "/dev/termination-log",
						"terminationMessagePolicy": "File",
						"imagePullPolicy": "Always"
					}],
					"restartPolicy": "Always",
					"terminationGracePeriodSeconds": 30,
					"dnsPolicy": "ClusterFirst",
					"serviceAccountName": "default",
					"serviceAccount": "default",
					"securityContext": {},
					"schedulerName": "default-scheduler",
					"tolerations": [{
						"key": "node.kubernetes.io/not-ready",
						"operator": "Exists",
						"effect": "NoExecute",
						"tolerationSeconds": 300
					}, {
						"key": "node.kubernetes.io/unreachable",
						"operator": "Exists",
						"effect": "NoExecute",
						"tolerationSeconds": 300
					}]
				},
				"status": {}
			},
			"oldObject": null
		}
	}`)
	req := httptest.NewRequest("POST", "/", admissionReview)

	rr := httptest.NewRecorder()
	EntryPoint(rr, req)

	out, err := ioutil.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	want := `{"response":{"uid":"test-no-nodeselector-present","allowed":true,"patch":"W3sib3AiOiJhZGQiLCJwYXRoIjoiL3NwZWMvbm9kZVNlbGVjdG9yIiwidmFsdWUiOnsicHVycG9zZSI6Im5vbi1wcm9kdWN0aW9uIn19XQ==","patchType":"JSONPatch"}}`
	if got := string(out); got != string(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEntryPointPOSTReplace(t *testing.T) {
	admissionReview := strings.NewReader(`{
		"kind": "AdmissionReview",
		"apiVersion": "admission.k8s.io/v1beta1",
		"request": {
			"uid": "test-nodeselector-present",
			"kind": {
				"group": "",
				"version": "v1",
				"kind": "Pod"
			},
			"resource": {
				"group": "",
				"version": "v1",
				"resource": "pods"
			},
			"namespace": "test",
			"name": "test",
			"operation": "CREATE",
			"userInfo": {
				"username": "go-test"
			},
			"object": {
				"metadata": {
					"generateName": "nginx-65899c769f-",
					"creationTimestamp": null,
					"labels": {
						"pod-template-hash": "2145573259",
						"run": "nginx"
					},
					"ownerReferences": [{
						"apiVersion": "extensions/v1beta1",
						"kind": "ReplicaSet",
						"name": "nginx-65899c769f",
						"uid": "56ff8242-cc6f-11e8-b891-42010a8e0fd6",
						"controller": true,
						"blockOwnerDeletion": true
					}]
				},
				"spec": {
					"volumes": [{
						"name": "default-token-wgwgs",
						"secret": {
							"secretName": "default-token-wgwgs"
						}
					}],
					"containers": [{
						"name": "nginx",
						"image": "nginx",
						"resources": {},
						"volumeMounts": [{
							"name": "default-token-wgwgs",
							"readOnly": true,
							"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
						}],
						"terminationMessagePath": "/dev/termination-log",
						"terminationMessagePolicy": "File",
						"imagePullPolicy": "Always"
					}],
					"restartPolicy": "Always",
					"terminationGracePeriodSeconds": 30,
					"dnsPolicy": "ClusterFirst",
					"serviceAccountName": "default",
					"serviceAccount": "default",
					"securityContext": {},
					"nodeSelector": {
						"purpose": "production"
					},
					"schedulerName": "default-scheduler",
					"tolerations": [{
						"key": "node.kubernetes.io/not-ready",
						"operator": "Exists",
						"effect": "NoExecute",
						"tolerationSeconds": 300
					}, {
						"key": "node.kubernetes.io/unreachable",
						"operator": "Exists",
						"effect": "NoExecute",
						"tolerationSeconds": 300
					}]
				},
				"status": {}
			},
			"oldObject": null
		}
	}`)
	req := httptest.NewRequest("POST", "/", admissionReview)

	rr := httptest.NewRecorder()
	EntryPoint(rr, req)

	out, err := ioutil.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	want := `{"response":{"uid":"test-nodeselector-present","allowed":true,"patch":"W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL25vZGVTZWxlY3RvciIsInZhbHVlIjp7InB1cnBvc2UiOiJub24tcHJvZHVjdGlvbiJ9fV0=","patchType":"JSONPatch"}}`
	if got := string(out); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
