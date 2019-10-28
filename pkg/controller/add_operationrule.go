package controller

import (
	"github.com/tchughesiv/inferator/pkg/controller/operationrule"
	"github.com/tchughesiv/inferator/pkg/controller/operationrule/logs"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logs.GetLogger("inferator.controller")

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	addManager := func(mgr manager.Manager) error {
		k8sService := GetInstance(mgr)
		reconciler := operationrule.Reconciler{Service: &k8sService}
		return operationrule.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
