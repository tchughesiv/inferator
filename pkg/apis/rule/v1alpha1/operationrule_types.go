package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperationRuleSpec defines the desired state of OperationRule
// +k8s:openapi-gen=true
type OperationRuleSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Resource OperationRuleSpecType `json:"resource,omitempty"`
}

// OperationRuleSpecType defines the desired state of OperationRule
// +k8s:openapi-gen=true
type OperationRuleSpecType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// OperationRuleStatus defines the observed state of OperationRule
// +k8s:openapi-gen=true
type OperationRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperationRule is the Schema for the operationrules API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=operationrules,scope=Namespaced
type OperationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperationRuleSpec   `json:"spec,omitempty"`
	Status OperationRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperationRuleList contains a list of OperationRule
type OperationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperationRule `json:"items"`
}

// PlatformService ...
type PlatformService interface {
	Create(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	List(ctx context.Context, list runtime.Object, opts client.ListOption) error
	Update(ctx context.Context, obj runtime.Object) error
	GetCached(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	GetDiscoveryClient() *discovery.DiscoveryClient
	GetScheme() *runtime.Scheme
	IsMockService() bool
}

func init() {
	SchemeBuilder.Register(&OperationRule{}, &OperationRuleList{})
}
