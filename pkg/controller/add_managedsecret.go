package controller

import (
	"github.com/linki/encrypted-secrets/pkg/controller/managedsecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, managedsecret.Add)
}
