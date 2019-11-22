package apis

import (
	"os"

	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	if os.Getenv(constants.RuntimeEnv) == "true" {
		schemeGroupVersion := schema.GroupVersion{Group: os.Getenv("OPRULE_OBJECT_GROUP"), Version: os.Getenv("OPRULE_OBJECT_VERSION")}
		schemeBuilder := &scheme.Builder{GroupVersion: schemeGroupVersion}
		AddToSchemes = append(AddToSchemes,
			v1alpha1.SchemeBuilder.AddToScheme,
			corev1.SchemeBuilder.AddToScheme,
			schemeBuilder.AddToScheme,
		)
	} else {
		AddToSchemes = append(AddToSchemes,
			v1alpha1.SchemeBuilder.AddToScheme,
			corev1.SchemeBuilder.AddToScheme,
			rbacv1.SchemeBuilder.AddToScheme,
		)
	}
}
