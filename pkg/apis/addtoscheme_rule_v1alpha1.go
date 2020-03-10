package apis

import (
	"os"

	// monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	knative "github.com/knative/serving/pkg/apis/serving/v1"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	configv1 "github.com/openshift/api/config/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	security1 "github.com/openshift/api/security/v1"
	ologgingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	csvv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/constants"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		ologgingv1.SchemeBuilder.AddToScheme,
		corev1.AddToScheme,
		appsv1.AddToScheme,
		rbacv1.AddToScheme,
		knative.AddToScheme,
		batchv1.AddToScheme,
		oappsv1.AddToScheme,
		security1.AddToScheme,
		routev1.AddToScheme,
		oimagev1.AddToScheme,
		buildv1.AddToScheme,
		configv1.AddToScheme,
		csvv1alpha1.AddToScheme,
	)
	if os.Getenv(constants.RuntimeEnv) == "true" {
		schemeGroupVersion := schema.GroupVersion{Group: os.Getenv("OPRULE_OBJECT_GROUP"), Version: os.Getenv("OPRULE_OBJECT_VERSION")}
		schemeBuilder := &scheme.Builder{GroupVersion: schemeGroupVersion}
		AddToSchemes = append(AddToSchemes,
			schemeBuilder.AddToScheme,
		)
	}
}
