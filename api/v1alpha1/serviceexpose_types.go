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

	// Backend networkingv1.IngressBackend https://github.com/kubernetes/kubernetes/blob/v1.21.1/pkg/apis/networking/types.go#L511-L524
	// +kubebuilder:validation:Required
	Backend networkingv1.IngressBackend `json:"backend"`

	// Path networkingv1.HTTPIngressPath.Path https://github.com/kubernetes/kubernetes/blob/v1.21.1/pkg/apis/networking/types.go#L493-L498
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`

	// PathType networkingv1.HTTPIngressPath.PathType https://github.com/kubernetes/kubernetes/blob/v1.21.1/pkg/apis/networking/types.go#L500-L504
	// +optional
	PathType networkingv1.PathType `json:"pathType,omitempty"`

	// Domain Host domain prefix generated in Ingress
	// eg SERVICE_NAME.NAMESPACE.Domain
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// IngressClassName
	// +optional
	IngressClassName string `json:"ingressClassName,omitempty"`

	// TLSEnabled Enable networkingv1.IngressSpec.TLS
	// https://github.com/kubernetes/kubernetes/blob/v1.21.1/pkg/apis/networking/types.go#L269-L276
	// +optional
	TLSEnabled bool `json:"tlsEnable,omitempty"`

	// TLSSecretName This secret name using networkingv1.IngressTLS.SecretName
	// https://github.com/kubernetes/kubernetes/blob/v1.21.1/pkg/apis/networking/types.go#L376-L382
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`

	// Annotations This annotation is generated in Ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ServiceExposeStatus defines the observed state of ServiceExpose
type ServiceExposeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// IngressName generated Ingress name
	IngressName string `json:"ingressName,omitempty"`

	// IngressName generated Ingress host
	IngressHost string `json:"ingressHost,omitempty"`

	// Ready Ingress generation status
	Ready corev1.ConditionStatus `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceExpose is the Schema for the serviceexposes API
//+kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`
//+kubebuilder:printcolumn:name="Ingress Name",type=string,JSONPath=`.status.ingressName`
//+kubebuilder:printcolumn:name="Ingress Host",type=string,JSONPath=`.status.ingressHost`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.ready`
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
