package controller

import (
	"github.com/j1cken/smooth-appdev-operator/pkg/controller/smoothupdate"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, smoothupdate.Add)
}
