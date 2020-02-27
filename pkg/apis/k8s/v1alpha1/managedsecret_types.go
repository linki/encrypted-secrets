package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ManagedSecretSpec defines the desired state of ManagedSecret
type ManagedSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:Enum=AWS;GCP
	Provider string `json:"provider"`
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`
}

// ManagedSecretStatus defines the observed state of ManagedSecret
type ManagedSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedSecret is the Schema for the managedsecrets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=managedsecrets,scope=Namespaced
type ManagedSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedSecretSpec   `json:"spec,omitempty"`
	Status ManagedSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedSecretList contains a list of ManagedSecret
type ManagedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedSecret{}, &ManagedSecretList{})
}