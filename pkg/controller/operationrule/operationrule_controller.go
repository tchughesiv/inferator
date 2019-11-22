package operationrule

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/constants"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/logs"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/admission"
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
	if os.Getenv(constants.RuntimeEnv) == "true" {
		gvk := schema.GroupVersionKind{Group: os.Getenv("OPRULE_OBJECT_GROUP"), Version: os.Getenv("OPRULE_OBJECT_VERSION"), Kind: os.Getenv("OPRULE_OBJECT_KIND")}
		objInt := admission.NewObjectInterfacesFromScheme(mgr.GetScheme())
		typer := objInt.GetObjectTyper()
		if typer.Recognizes(gvk) {
			creator := objInt.GetObjectCreater()
			newObject, err := creator.New(gvk)
			if err != nil {
				return err
			}
			log.Info(newObject.GetObjectKind().GroupVersionKind().String())
			if err != nil {
				return err
			}
			watchObjects := []runtime.Object{
				newObject,
			}
			objectHandler := &handler.EnqueueRequestForObject{}
			for _, watchObject := range watchObjects {
				err = c.Watch(&source.Kind{Type: watchObject}, objectHandler)
				if err != nil {
					return err
				}
			}
		}
	} else {
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
	if os.Getenv(constants.RuntimeEnv) == "true" {
		return r.ReconcileInferator(request)
	}
	return r.ReconcileOperator(request)
}

func (r *Reconciler) ReconcileInferator(request reconcile.Request) (reconcile.Result, error) {
	if request.Name == os.Getenv("OPRULE_OBJECT_NAME") {
		log := log.With("Inferator", request.Name, "Namespace", request.Namespace)
		log.Info("OPRULE_OBJECT_NAME = " + os.Getenv("OPRULE_OBJECT_NAME"))
		log.Info("OPRULE_OBJECT_KIND = " + os.Getenv("OPRULE_OBJECT_KIND"))
		log.Info("OPRULE_OBJECT_GROUP_VERSION = " + os.Getenv("OPRULE_OBJECT_GROUP_VERSION"))
		log.Info("OPRULE_SCHEMA_NAME = " + os.Getenv("OPRULE_SCHEMA_NAME"))

		gvk := schema.GroupVersionKind{Group: os.Getenv("OPRULE_OBJECT_GROUP"), Version: os.Getenv("OPRULE_OBJECT_VERSION"), Kind: os.Getenv("OPRULE_OBJECT_KIND")}
		objInt := admission.NewObjectInterfacesFromScheme(r.Service.GetScheme())
		typer := objInt.GetObjectTyper()
		if typer.Recognizes(gvk) {
			creator := objInt.GetObjectCreater()
			object, err := creator.New(gvk)
			if err != nil {
				return reconcile.Result{}, err
			}

			err = r.Service.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, object)
			if err != nil {
				return reconcile.Result{}, err
			}
			prettyJSON, err := json.MarshalIndent(object, "", "    ")
			if err != nil {
				return reconcile.Result{}, err
			}
			fmt.Printf("%s\n", string(prettyJSON))
		}
	}

	time.Sleep(60 * time.Second)
	return reconcile.Result{}, nil
}

func (r *Reconciler) ReconcileOperator(request reconcile.Request) (reconcile.Result, error) {
	log := log.With("OperationRule", request.Name, "Namespace", request.Namespace)

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
	} else if err != nil {
		return reconcile.Result{}, err
	}

	schemaName := instance.Spec.Resource.GroupVersionKind().Group + "." +
		instance.Spec.Resource.GroupVersionKind().Version + "." +
		instance.Spec.Resource.Kind

	d, _ := r.Service.GetDiscoveryClient().OpenAPISchema()
	for _, x := range d.Definitions.AdditionalProperties {
		if strings.HasSuffix(x.Name, schemaName) {
			schemaName = x.Name
		}
	}

	fmt.Println("NamedSchema: " + schemaName)

	// Define a new Pod object
	pod := newPodForCR(instance, serviceAccount.Name, schemaName, namespace)
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
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *rulev1alpha1.OperationRule, serviceAccount, schemaName, namespace string) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	name := cr.Name + "-inferator"
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
			Containers: []corev1.Container{
				{
					Name:    name,
					Image:   "quay.io/tchughesiv/inferator",
					Command: []string{"inferator"},
					Env: []corev1.EnvVar{
						{Name: constants.RuntimeEnv, Value: "true"},
						{Name: "OPRULE_SCHEMA_NAME", Value: schemaName},
						{Name: "OPRULE_OBJECT_NAME", Value: cr.Spec.Resource.Name},
						{Name: "OPRULE_OBJECT_KIND", Value: cr.Spec.Resource.Kind},
						{Name: "OPRULE_OBJECT_GROUP", Value: cr.Spec.Resource.GroupVersionKind().Group},
						{Name: "OPRULE_OBJECT_VERSION", Value: cr.Spec.Resource.GroupVersionKind().Version},
						{Name: "OPRULE_OBJECT_GROUP_VERSION", Value: cr.Spec.Resource.GroupVersionKind().GroupVersion().String()},
						{Name: "WATCH_NAMESPACE", Value: namespace},
						{Name: "POD_NAME", Value: name},
						{Name: "OPERATOR_NAME", Value: name},
					},
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
			{
				APIGroups: []string{"", "rbac.authorization.k8s.io", v1alpha1.SchemeGroupVersion.Group},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "services", "pods", "pods/finalizers"},
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
