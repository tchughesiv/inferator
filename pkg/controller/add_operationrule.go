package controller

import (
	"github.com/tchughesiv/inferator/pkg/controller/operationrule"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, operationrule.Add)
}
