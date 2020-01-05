// Package dummy exists only for technical reasons, to ensure that command-line
// tools needed for `make generate` to work will get downloaded into `vendor/`
// directory.
package dummy

import (
	client_gen "k8s.io/code-generator/cmd/client-gen/args"
	deepcopy_gen "k8s.io/code-generator/cmd/deepcopy-gen/args"
	informer_gen "k8s.io/code-generator/cmd/informer-gen/args"
	lister_gen "k8s.io/code-generator/cmd/lister-gen/args"
	openapi_gen "k8s.io/kube-openapi/cmd/openapi-gen/args"
)

var _ = client_gen.Validate
var _ = deepcopy_gen.Validate
var _ = informer_gen.Validate
var _ = lister_gen.Validate
var _ = openapi_gen.Validate
