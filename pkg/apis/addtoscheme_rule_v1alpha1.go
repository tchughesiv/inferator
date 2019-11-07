package apis

import (
	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme, appsv1.SchemeBuilder.AddToScheme)
}
