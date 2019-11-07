package apis

import (
	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		corev1.SchemeBuilder.AddToScheme,
		rbacv1.SchemeBuilder.AddToScheme,
	)
}
