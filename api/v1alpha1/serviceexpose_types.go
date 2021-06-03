/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceExposeSpec defines the desired state of ServiceExpose
type ServiceExposeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Backend
	// +kubebuiler:validation:Required
	Backend networkingv1.IngressBackend `json:"backend"`

	// Path
	// +kubebuiler:validation:Required
	Path string `json:"path,omitempty"`

	// PathType
	// +optional
	PathType networkingv1.PathType `json:"pathType,omitempty"`

	// Domain
	// +kubebuiler:validation:Required
	Domain string `json:"domain"`

	// TLSEnabled
	// +optional
	TLSEnabled bool `json:"tls_enable,omitempty"`

	// TLSSecretName
	// +optional
	TLSSecretName string `json:"tls_secret_name,omitempty"`

	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ServiceExposeStatus defines the observed state of ServiceExpose
type ServiceExposeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	IngressName string                 `json:"ingress_name,omitempty"`
	IngressHost string                 `json:"ingress_host,omitempty"`
	Revision    string                 `json:"revision"`
	Ready       corev1.ConditionStatus `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceExpose is the Schema for the serviceexposes API
type ServiceExpose struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceExposeSpec   `json:"spec,omitempty"`
	Status ServiceExposeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceExposeList contains a list of ServiceExpose
type ServiceExposeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceExpose `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceExpose{}, &ServiceExposeList{})
}
