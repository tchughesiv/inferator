package operationrule

import (
	"context"

	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/logs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logs.GetLogger("controller_operationrule")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new OperationRule Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	discoveryclient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Error getting image client.")
		return &ReconcileOperationRule{}
	}
	return &ReconcileOperationRule{client: mgr.GetClient(), discoveryclient: discoveryclient, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("operationrule-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	watchObjects := []runtime.Object{
		&rulev1alpha1.OperationRule{},
	}
	objectHandler := &handler.EnqueueRequestForObject{}
	for _, watchObject := range watchObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, objectHandler)
		if err != nil {
			return err
		}
	}

	watchOwnedObjects := []runtime.Object{
		&appsv1.Deployment{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rulev1alpha1.OperationRule{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileOperationRule implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOperationRule{}

// ReconcileOperationRule reconciles a OperationRule object
type ReconcileOperationRule struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	discoveryclient *discovery.DiscoveryClient
	scheme          *runtime.Scheme
}

// Reconcile reads that state of the cluster for a OperationRule object and makes changes based on the state read
// and what is in the OperationRule.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOperationRule) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := log.With("OperationRule", request.Name, "Namespace", request.Namespace)
	log.Info("Reconciling")

	// Fetch the OperationRule instance
	instance := &rulev1alpha1.OperationRule{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	apiResourceList, err := r.discoveryclient.ServerResourcesForGroupVersion(instance.Spec.Resource.GroupVersionKind().GroupVersion().String())
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error("GroupVersion ", instance.Spec.Resource.GroupVersionKind().GroupVersion(), " does not exist in the cluster.")
		}
		return reconcile.Result{}, nil
	}
	exists := false
	for _, resource := range apiResourceList.APIResources {
		if resource.Kind == instance.Spec.Resource.Kind {
			exists = true
			log.Info(resource.String())
			break
		}
	}
	if !exists {
		log.Error("Kind ", instance.Spec.Resource.Kind, " does not exist for ", instance.Spec.Resource.GroupVersionKind().GroupVersion(), " in the cluster.")
		return reconcile.Result{}, nil
	}

	namespace := instance.Namespace
	if instance.Spec.Resource.Namespace != "" {
		namespace = instance.Spec.Resource.Namespace
	}

	// Define a new Pod object
	pod := newPodForCR(instance, namespace)

	// Set OperationRule instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Pod in ", "Namespace ", namespace, " Pod.Name ", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	log.Info("Skip reconcile: Pod already exists in ", "Namespace ", found.Namespace, " Pod.Name ", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *rulev1alpha1.OperationRule, namespace string) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
