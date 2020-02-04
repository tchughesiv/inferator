package operationrule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	knative "github.com/knative/serving/pkg/apis/serving/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/components"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/constants"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/logs"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apiserver/pkg/admission"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logs.GetLogger("controller_operationrule")

//KubeObject ...
type KubeObject interface {
	runtime.Object
	metav1.Object
}

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
		objects := map[string]rulev1alpha1.OperationRuleSpecType{}
		watchObjects := []runtime.Object{}
		err = json.Unmarshal([]byte(os.Getenv("OPRULE_OBJECTS")), &objects)
		if err != nil {
			return err
		}
		for _, s := range objects {
			objInt := admission.NewObjectInterfacesFromScheme(mgr.GetScheme())
			creator := objInt.GetObjectCreater()
			newObject, err := creator.New(s.GroupVersionKind())
			if err != nil {
				return err
			}
			log.Info(newObject.GetObjectKind().GroupVersionKind().String())
			if err != nil {
				return err
			}
			watchObjects = append(watchObjects, newObject)
		}

		objectHandler := &handler.EnqueueRequestForObject{}
		for _, watchObject := range watchObjects {
			err = c.Watch(&source.Kind{Type: watchObject}, objectHandler)
			if err != nil {
				return err
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
			&corev1.Service{},
			&corev1.ServiceAccount{},
			&rbacv1.Role{},
			&rbacv1.RoleBinding{},
			// &knative.Service{},
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
		return r.reconcileInferator(request)
	}
	return r.reconcileOperator(request)
}

func itemExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

// reconcileInferator ...
func (r *Reconciler) reconcileInferator(request reconcile.Request) (reconcile.Result, error) {
	log := log.With("Inferator", request.Name, "Namespace", request.Namespace)
	inputs := []string{}
	err := json.Unmarshal([]byte(os.Getenv("OPRULE_INPUTS")), &inputs)
	if err != nil {
		log.Error(err)
		return reconcile.Result{}, err
	}
	objects := map[string]rulev1alpha1.OperationRuleSpecType{}
	err = json.Unmarshal([]byte(os.Getenv("OPRULE_OBJECTS")), &objects)
	if err != nil {
		log.Error(err)
		return reconcile.Result{}, err
	}
	postObject := map[string]runtime.Object{}
	objInt := admission.NewObjectInterfacesFromScheme(r.Service.GetScheme())
	creator := objInt.GetObjectCreater()
	objectNames := []string{}

	// only reconcile resources of interest
	for name, obj := range objects {
		if obj.Namespace == "" {
			obj.Namespace = request.Namespace
			objects[name] = obj
		}
		if itemExists(inputs, name) {
			object, err := creator.New(obj.GroupVersionKind())
			if err != nil {
				log.Error(err)
			}
			objectNames = append(objectNames, obj.Name)
			postObject[name] = object
		}
	}

	var post bool
	if itemExists(objectNames, request.Name) {
		// only send 'input' objects to zenithr
		for name, object := range postObject {
			err = r.Service.Get(context.TODO(), types.NamespacedName{Name: objects[name].Name, Namespace: objects[name].Namespace}, object)
			if err != nil {
				log.Error(err)
			} else {
				if objects[name].Name == request.Name {
					println()
					log.Infof("Call zenithr service for %s %s", object.GetObjectKind().GroupVersionKind().Kind, request.Name)
					println()
				}
				post = true
				postObject[name] = object
			}
		}
	}

	if post {
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(&postObject)
		if err != nil {
			log.Error(err)
		}

		resp, err := http.Post("http://localhost:8080/", "application/json", &buf)
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
		}

		variables := []rulev1alpha1.Variable{}
		err = json.Unmarshal(body, &variables)
		if err != nil {
			log.Error(err)
		}

		prettyJSONresp, err := json.MarshalIndent(variables, "", "    ")
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("%s\n", string(prettyJSONresp))

		for _, v := range variables {
			var update bool
			obj := objects[v.Name]
			//	if objInt.GetObjectTyper().Recognizes(obj.GetObjectKind().GroupVersionKind()) {
			object, err := creator.New(obj.GroupVersionKind())
			if err != nil {
				log.Error(err)
			}
			err = r.Service.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, object)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Warn(obj.Kind + " " + obj.Name + " for " + v.Name + " was not found")
				} else {
					log.Error(err)
				}
			} else {
				update = true
			}

			if update {
				existingJSON, err := json.Marshal(&object)
				if err != nil {
					log.Error("Unmarshal " + v.Name)
				}
				newJSON := fieldTypeConversion(v, object)
				if !reflect.DeepEqual(existingJSON, newJSON) {
					objectOut, err := creator.New(obj.GroupVersionKind())
					if err != nil {
						log.Error(err)
					}
					if err = json.Unmarshal(newJSON, &objectOut); err != nil {
						if ute, ok := err.(*json.UnmarshalTypeError); ok {
							rval := reflect.Zero(ute.Type)
							fmt.Println("Unmarshal error - should be type '" + rval.Kind().String() + "'")
						} else {
							log.Error(err)
						}
					}
					log.Infof("Updating %s %s", objectOut.GetObjectKind().GroupVersionKind().Kind, objects[v.Name].Name)
					err = r.Service.Update(context.TODO(), objectOut)
					if err != nil {
						log.Error(err)
					}
				}
			}
		}
	}

	return reconcile.Result{}, nil
}

func fieldTypeConversion(v rulev1alpha1.Variable, object runtime.Object) []byte {
	var err error
	existingJSON, err := json.Marshal(object)
	if err != nil {
		log.Error(err)
	}
	newJSON := existingJSON
	gResult := gjson.GetBytes(newJSON, v.Path)
	if gResult.IsObject() {
		gval := reflect.ValueOf(gResult.Value())
		for s, val := range v.Value {
			gNested := gjson.Get(gResult.String(), s)
			rval := reflect.ValueOf(gNested.Value())
			switch rval.Kind() {
			default:
				new := val
				if gNested.Value() != val {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err = json.Unmarshal(newJSON, &object); err != nil {
						if ute, ok := err.(*json.UnmarshalTypeError); ok {
							rval = reflect.Zero(ute.Type)
						} else {
							log.Error(err)
						}
					}
				}
				fallthrough
			case reflect.Bool:
				new, err := strconv.ParseBool(val)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Int:
				new, err := strconv.ParseInt(val, 10, 0)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Int8:
				new, err := strconv.ParseInt(val, 10, 8)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Int32:
				new, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Int64:
				new, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
				fmt.Printf("int: %v\n", rval.Uint())
				new, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Float32:
				new, err := strconv.ParseFloat(val, 32)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Float64:
				new, err := strconv.ParseFloat(val, 64)
				if err != nil {
					log.Error(rval.Kind().String() + " conversion failed")
				}
				if gNested.Value() != new {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.String:
				new := val
				if gNested.Value() != val {
					newJSON, err = sjson.SetBytes(newJSON, v.Path+"."+s, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Map:
				new := map[string]string{}
				tmpJSON, err := json.Marshal(gval.Interface())
				if err != nil {
					log.Error(err)
				}
				err = json.Unmarshal(tmpJSON, &new)
				if err != nil {
					log.Error(err)
				}
				for key, value := range v.Value {
					new[key] = value
				}
				if gNested.Value() != gval.Interface() {
					newJSON, err = sjson.SetBytes(newJSON, v.Path, new)
					if err != nil {
						log.Error(err)
					}
				}
			case reflect.Slice:
				fmt.Printf("slice: len=%d, %v\n", rval.Len(), rval.Interface())
			case reflect.Chan:
				fmt.Printf("chan %v\n", rval.Interface())
			}
		}
	}
	return newJSON
}

// reconcileOperator ...
func (r *Reconciler) reconcileOperator(request reconcile.Request) (reconcile.Result, error) {
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

	for _, i := range instance.Spec.Inference.Inputs {
		if _, ok := instance.Spec.Resources[i]; !ok {
			log.Error(i + " is not declared as a resource")
			return reconcile.Result{}, nil
		}
	}
	for _, rules := range instance.Spec.Inference.Rules {
		for _, output := range rules.Then {
			if _, ok := instance.Spec.Resources[output.Name]; !ok {
				log.Error(output.Name + " is not declared as a resource")
				return reconcile.Result{}, nil
			}
		}
	}
	namespace := instance.Namespace
	resources := []rulev1alpha1.OperationRuleSpecType{}
	apiResources := []metav1.APIResource{}
	for _, resource := range instance.Spec.Resources {
		apiResourceList, err := r.Service.GetDiscoveryClient().ServerResourcesForGroupVersion(resource.GroupVersionKind().GroupVersion().String())
		if err != nil {
			if errors.IsNotFound(err) {
				log.Error("GroupVersion ", resource.GroupVersionKind().GroupVersion(), " does not exist in the cluster.")
			}
			return reconcile.Result{}, nil
		}

		exists := false
		for _, apiR := range apiResourceList.APIResources {
			if resource.Kind == apiR.Kind {
				exists = true
				// append to array for later processing
				resources = append(resources, resource)
				apiResources = append(apiResources, apiR)
				break
			}
		}
		if !exists {
			log.Error("Kind ", resource.Kind, " does not exist for ", resource.GroupVersionKind().GroupVersion(), " in the cluster.")
			return reconcile.Result{}, nil
		}

		if resource.Namespace != "" {
			namespace = resource.Namespace
		}

		/*
			schemaName := resource.GroupVersionKind().Group + "." +
				resource.GroupVersionKind().Version + "." +
				resource.Kind

			d, _ := r.Service.GetDiscoveryClient().OpenAPISchema()
			for _, x := range d.Definitions.AdditionalProperties {
				if strings.HasSuffix(x.Name, schemaName) {
					schemaName = x.Name
				}
			}

			println("NamedSchema: " + schemaName)
		*/
	}
	// Define a new Role object
	role := newRoleforCR(instance, resources, apiResources, namespace)
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

	// Define a new Pod object
	/*
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
	*/

	///////// START ////////

	if instance.Spec.KNative {
		genKService := newKService(instance)
		curKService := &knative.Service{}
		err = r.loadOrCreate(instance, genKService, curKService)
		if err != nil {
			return reconcile.Result{}, err
		} else if existed(curKService) {
			curContainer := curKService.Spec.ConfigurationSpec.Template.Spec.Containers[1]
			genContainer := genKService.Spec.ConfigurationSpec.Template.Spec.Containers[1]
			updated, err := changed(curContainer, genContainer)
			if err != nil {
				log.Info("Detected that knative service remains unchanged")
				return reconcile.Result{}, err
			} else if updated {
				log.Info("Detected that knative service needs to be updated")
				genKService.SetResourceVersion(curKService.GetResourceVersion())
				err = r.Service.Update(context.TODO(), genKService)
				if err != nil {
					return reconcile.Result{}, err
				}
				return reconcile.Result{Requeue: true}, nil
			} else {
				if instance.Status.RouteHost != getHostname(curKService.Status.URL.Host) {
					log.Info("Will set hostname to", "hostname", curKService.Status.URL.Host)
					instance.Status.RouteHost = getHostname(curKService.Status.URL.Host)
					err = r.Service.Update(context.TODO(), instance)
					if err != nil {
						log.Error(err, "Error updating CR", "cr", instance)
						return reconcile.Result{}, err
					}
				}

			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Create pod object based on CR, if does not exist:
	csv := components.Csv
	genPod := newPodForCR(instance, instance.Spec.Resources, serviceAccount.Name, namespace, csv.Registry, csv.Context, csv.ImageName, csv.Tag)
	curPod := &corev1.Pod{}
	err = r.loadOrCreate(instance, genPod, curPod)
	if err != nil {
		return reconcile.Result{}, err
	} else if existed(curPod) {
		curContainer := curPod.Spec.Containers[1]
		genContainer := genPod.Spec.Containers[1]
		updated, err := changed(curContainer, genContainer)
		if err != nil {
			return reconcile.Result{}, err
		} else if updated {
			log.Info("Detected that pod needs to be updated, will delete it and let it be recreated!")
			err = r.Service.Delete(context.TODO(), curPod)
			if err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}

	genRoute := newRouteForCR(instance)
	curRoute := &routev1.Route{}
	if instance.Spec.Expose {
		// Create service object based on CR, if does not exist:
		curService := &corev1.Service{}
		err = r.loadOrCreate(instance, newServiceForCR(instance), curService)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Create route based on CR, if does not exist:
		err = r.loadOrCreate(instance, genRoute, curRoute)
		if err != nil {
			return reconcile.Result{}, err
		} else if existed(curRoute) {
			if len(instance.Spec.HostName) > 0 && instance.Spec.HostName != curRoute.Spec.Host {
				log.Info("Detected that route hostname needs to be updated!")
				curRoute.Spec.Host = instance.Spec.HostName
				err = r.Service.Update(context.TODO(), curRoute)
				if err != nil {
					return reconcile.Result{}, err
				}
				//Status URL should next be updated based on this
				return reconcile.Result{Requeue: true}, nil
			}
		}
	} else {
		err := r.Service.Get(context.TODO(), types.NamespacedName{Name: genRoute.Name, Namespace: genRoute.Namespace}, curRoute)
		if err == nil {
			//There is a route from before, but delete it, since expose flag has been removed
			log.Info("Will delete old route")
			err = r.Service.Delete(context.TODO(), curRoute)
			if err != nil {
				log.Info("Error deleting", "error", err)
				return reconcile.Result{}, err
			}
		} else if errors.IsNotFound(err) {
			//There is no existing route, nor should there be one, so all is good
		} else {
			//Unknown error
			log.Info("Error finding out if there was an old route", "error", err)
			return reconcile.Result{}, err
		}
	}
	if instance.Spec.Expose {
		if len(curRoute.Name) == 0 {
			//Route must have been just created, let's set URL status later
			retryTime := 5
			log.Info("Will try reconciliation again to set status hostname", "retry time", retryTime)
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(retryTime) * time.Second}, nil
		}
		if instance.Status.RouteHost != getHostname(curRoute.Spec.Host) {
			err := r.setRouteHostname(instance, *curRoute)
			if err != nil {
				log.Error(err, "Error setting route hostname")
				return reconcile.Result{}, err
			}
			retryTime := 5
			log.Info("Should have updated route host, but will try reconciliation again to verify", "retry time", retryTime)
			return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(retryTime) * time.Second}, nil
		}
	} else if len(instance.Status.RouteHost) > 0 {
		instance.Status.RouteHost = ""
		err = r.Service.Update(context.TODO(), instance)
		if err != nil {
			log.Error(err, "Error updating CR", "cr", instance)
			return reconcile.Result{}, err
		}
	}
	///////// STOP ////////

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *rulev1alpha1.OperationRule, resources map[string]rulev1alpha1.OperationRuleSpecType, serviceAccount, namespace, repository, context, imageName, tag string) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	registryName := strings.Join([]string{repository, context, imageName}, "/")
	image := strings.Join([]string{registryName, tag}, ":")
	name := cr.Name + "-inferator"
	zname := cr.Name + "-zenithr"
	inputs, err := json.Marshal(cr.Spec.Inference.Inputs)
	if err != nil {
		log.Error(err)
	}
	objects, err := json.Marshal(resources)
	if err != nil {
		log.Error(err)
	}
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
					Name:            name,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"inferator"},
					Env: []corev1.EnvVar{
						{Name: constants.RuntimeEnv, Value: "true"},
						{Name: "OPRULE_OBJECTS", Value: string(objects)},
						{Name: "OPRULE_INPUTS", Value: string(inputs)},
						{Name: "WATCH_NAMESPACE", Value: namespace},
						{Name: "POD_NAME", Value: name},
						{Name: "OPERATOR_NAME", Value: name},
					},
				},
				{
					Name:            zname,
					Image:           "docker.io/ruromero/zenithr-service-jdk8",
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  constants.RulesVar,
							Value: getJSON(cr.Spec.Inference),
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "health",
								Port: intstr.IntOrString{IntVal: 8080},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						FailureThreshold:    20,
					},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "health",
								Port: intstr.IntOrString{IntVal: 8080},
							},
						},
						InitialDelaySeconds: 60,
						PeriodSeconds:       60,
					},
				},
			},
		},
	}
}

// newRoleforCR ...
func newRoleforCR(cr *rulev1alpha1.OperationRule, resources []rulev1alpha1.OperationRuleSpecType, apiResources []metav1.APIResource, namespace string) *rbacv1.Role {
	labels := map[string]string{
		"app": cr.Name,
	}
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "rbac.authorization.k8s.io", v1alpha1.SchemeGroupVersion.Group},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "services", "pods", "pods/finalizers"},
				Verbs: []string{
					"create",
					"delete",
					"deletecollection",
					"get",
					"list",
					"patch",
					"update",
					"watch",
				},
			},
		},
	}
	for i, apiResource := range apiResources {
		role.Rules = append(role.Rules,
			rbacv1.PolicyRule{
				APIGroups: []string{resources[i].GroupVersionKind().Group},
				Resources: []string{apiResource.Name},
				Verbs:     apiResource.Verbs,
			},
		)
	}
	return role
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

// newServiceForCR returns a service that directs to the application pod
func newServiceForCR(cr *rulev1alpha1.OperationRule) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 8080,
				},
			},
			Selector: labels,
		},
	}
}

// newRouteForCR returns a route that exposes the application service
func newRouteForCR(cr *rulev1alpha1.OperationRule) *routev1.Route {
	labels := map[string]string{
		"app": cr.Name,
	}
	route := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Name: cr.Name,
			},
		},
	}
	if len(cr.Spec.HostName) > 0 {
		route.Spec.Host = cr.Spec.HostName
	}
	route.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
	return &route
}

func newKService(cr *rulev1alpha1.OperationRule) *knative.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	klabels := map[string]string{
		"knative.dev/type": "container",
	}
	service := knative.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: knative.ServiceSpec{
			ConfigurationSpec: knative.ConfigurationSpec{
				Template: knative.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: klabels,
					},
					Spec: knative.RevisionSpec{},
				},
			},
		},
	}
	service.Spec.ConfigurationSpec.Template.Spec.Containers = []corev1.Container{
		{
			Image:           "docker.io/ruromero/zenithr-service-jdk8",
			ImagePullPolicy: corev1.PullAlways,
			Env: []corev1.EnvVar{
				{
					Name:  constants.RulesVar,
					Value: getJSON(cr.Spec.Inference),
				},
			},
		},
	}

	service.SetGroupVersionKind(knative.SchemeGroupVersion.WithKind("Service"))
	return &service
}

func getJSON(spec rulev1alpha1.OperationRuleSpecInference) string {
	bytes, err := json.Marshal(spec)
	if err != nil {
		panic("Failed to parse input!")
	}
	return string(bytes)
}

func (r *Reconciler) setRouteHostname(cr *rulev1alpha1.OperationRule, route routev1.Route) (err error) {
	hostname := getHostname(route.Spec.Host)
	if len(hostname) > 0 {
		err = r.Service.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
		if err != nil {
			log.Error(err, "Error Reloading CR", "cr", cr)
			return
		}
		log.Info("Will set route hostname to", "hostname", hostname)
		cr.Status.RouteHost = hostname
		err = r.Service.Update(context.TODO(), cr)
		if err != nil {
			log.Error(err, "Error updating CR", "cr", cr)
			return
		}
	}
	return
}

func getHostname(routeHost string) string {
	if len(routeHost) > 0 {
		return fmt.Sprintf("http://%s", routeHost)
	}
	return ""
}

func (r *Reconciler) loadOrCreate(instance *rulev1alpha1.OperationRule, genObject KubeObject, curObject KubeObject) error {
	log := log.With("Request.Namespace", instance.Namespace, "Request.Name", instance.Name)
	// Set DecisionService instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, genObject, r.Service.GetScheme()); err != nil {
		return err
	}
	//Check if this object already exists
	err := r.Service.Get(context.TODO(), types.NamespacedName{Name: genObject.GetName(), Namespace: genObject.GetNamespace()}, curObject)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Object ", "type ", reflect.TypeOf(genObject), " ", genObject.GetName())
		err = r.Service.Create(context.TODO(), genObject)
		if err != nil {
			log.Error("Got an error creating it", "error", err)
			return err
		}
		// Object created successfully - don't requeue
		return nil
	} else if err != nil {
		log.Error("Got an error looking it up", "error", err)
		return err
	} else {
		return nil
	}
}

func changed(current corev1.Container, generated corev1.Container) (changed bool, err error) {
	currentRules := getEnvVar(current.Env, constants.RulesVar)
	var currentSpec rulev1alpha1.OperationRuleSpec
	err = json.Unmarshal([]byte(currentRules), &currentSpec)
	if err != nil {
		return
	}
	generatedRules := getEnvVar(generated.Env, constants.RulesVar)
	var generatedSpec rulev1alpha1.OperationRuleSpec
	err = json.Unmarshal([]byte(generatedRules), &generatedSpec)
	if err != nil {
		return
	}
	if !reflect.DeepEqual(currentSpec, generatedSpec) {
		changed = true
	}
	return
}

func getEnvVar(vars []corev1.EnvVar, key string) string {
	for _, envVar := range vars {
		if envVar.Name == key {
			return envVar.Value
		}
	}
	return ""
}

func existed(object KubeObject) bool {
	return len(object.GetName()) > 0
}
