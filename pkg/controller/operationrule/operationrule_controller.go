package operationrule

import (
	"context"

	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/logs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
func Add(mgr manager.Manager, reconciler *Reconciler) error {
	return add(mgr, reconciler)
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
		&corev1.Pod{},
		&corev1.ServiceAccount{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
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
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a OperationRule object
type Reconciler struct {
	Service rulev1alpha1.PlatformService
}

// Reconcile reads that state of the cluster for a OperationRule object and makes changes based on the state read
// and what is in the OperationRule.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := log.With("OperationRule", request.Name, "Namespace", request.Namespace)
	log.Info("Reconciling")

	// Fetch the OperationRule instance
	instance := &rulev1alpha1.OperationRule{}
	err := r.Service.Get(context.TODO(), request.NamespacedName, instance)
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

	apiResourceList, err := r.Service.GetDiscoveryClient().ServerResourcesForGroupVersion(instance.Spec.Resource.GroupVersionKind().GroupVersion().String())
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error("GroupVersion ", instance.Spec.Resource.GroupVersionKind().GroupVersion(), " does not exist in the cluster.")
		}
		return reconcile.Result{}, nil
	}

	exists := false
	apiResource := metav1.APIResource{}
	for _, resource := range apiResourceList.APIResources {
		if resource.Kind == instance.Spec.Resource.Kind {
			exists = true
			apiResource = resource
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

	// Define a new Role object
	role := newRoleforCR(instance, apiResource, namespace)
	// Set OperationRule instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, role, r.Service.GetScheme()); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Role already exists
	foundRole := &rbacv1.Role{}
	err = r.Service.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Role in Namespace ", role.Namespace, " Name ", role.Name)
		err = r.Service.Create(context.TODO(), role)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Role created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Role object
	serviceAccount := newSAforCR(instance, namespace)
	// Set OperationRule instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceAccount, r.Service.GetScheme()); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Role already exists
	foundSA := &corev1.ServiceAccount{}
	err = r.Service.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundSA)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new ServiceAccount in Namespace ", serviceAccount.Namespace, " Name ", serviceAccount.Name)
		err = r.Service.Create(context.TODO(), serviceAccount)
		if err != nil {
			return reconcile.Result{}, err
		}

		// ServiceAccount created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Role object
	roleBinding := newRoleBindingforCR(instance, role.Name, serviceAccount.Name, namespace)
	// Set OperationRule instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, roleBinding, r.Service.GetScheme()); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Role already exists
	foundRoleBinding := &rbacv1.RoleBinding{}
	err = r.Service.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, foundRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new RoleBinding in Namespace ", roleBinding.Namespace, " Name ", roleBinding.Name)
		err = r.Service.Create(context.TODO(), roleBinding)
		if err != nil {
			return reconcile.Result{}, err
		}

		// RoleBinding created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := newPodForCR(instance, serviceAccount.Name, namespace)
	// Set OperationRule instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.Service.GetScheme()); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Pod already exists
	foundPod := &corev1.Pod{}
	err = r.Service.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Pod in ", "Namespace ", pod.Namespace, " Pod.Name ", pod.Name)
		err = r.Service.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	//schemaDoc, _ := r.Service.GetDiscoveryClient()

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *rulev1alpha1.OperationRule, serviceAccount, namespace string) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
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

// newRoleforCR ...
func newRoleforCR(cr *rulev1alpha1.OperationRule, apiResource metav1.APIResource, namespace string) *rbacv1.Role {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{cr.Spec.Resource.GroupVersionKind().Group},
				Resources: []string{apiResource.Name},
				Verbs:     apiResource.Verbs,
			},
		},
	}
}

// newRoleBindingforCR ...
func newRoleBindingforCR(cr *rulev1alpha1.OperationRule, roleName, serviceAccount, namespace string) *rbacv1.RoleBinding {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: rbacv1.ServiceAccountKind,
				Name: serviceAccount,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: roleName,
		},
	}
}

// newSAforCR ...
func newSAforCR(cr *rulev1alpha1.OperationRule, namespace string) *corev1.ServiceAccount {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
