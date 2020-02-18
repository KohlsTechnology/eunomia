package util

import "k8s.io/apimachinery/pkg/types"

// NN is a shorthand alias for types.NamespacedName. It is intended to help
// reduce "visual stutter" when specifying NamespacedName structs, while
// retaining all the information. For example, function calls that require a
// types.NamespacedName can now be written in a way somewhat resembling "named
// arguments" syntax from other programming languages:
//
//    someFunc(..., util.NN{Name: "foobar", Namespace: "barfoo"}, ...)
//
// compared to sample code it's intended to replace:
//
//    someFunc(..., types.NamespacedName{Name: "foobar", Namespace: "barfoo"}, ...)
type NN = types.NamespacedName

// NameNamespacer is an interface for objects which provide functions
// retrieving name and namespace in a way common in kubernetes API.
type NameNamespacer interface {
	GetNamespace() string
	GetName() string
}

// GetNN retrieves name & namespace from obj and wraps them in an NN object.
func GetNN(obj NameNamespacer) NN {
	return NN{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}
