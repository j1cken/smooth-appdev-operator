package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SmoothUpdateSpec defines the desired state of SmoothUpdate
// +k8s:openapi-gen=true
type SmoothUpdateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Deployment string `json:"deployment"`
	Version    string `json:"version"`
	UpdateSQL  string `json:"sql"`
}

// SmoothUpdateStatus defines the observed state of SmoothUpdate
// +k8s:openapi-gen=true
type SmoothUpdateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SmoothUpdate is the Schema for the smoothupdates API
// +k8s:openapi-gen=true
type SmoothUpdate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SmoothUpdateSpec   `json:"spec,omitempty"`
	Status SmoothUpdateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SmoothUpdateList contains a list of SmoothUpdate
type SmoothUpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SmoothUpdate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SmoothUpdate{}, &SmoothUpdateList{})
}
