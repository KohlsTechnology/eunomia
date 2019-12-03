package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitConfig represents all the infomration necessary to
type GitConfig struct {
	//+kubebuilder:validation:Pattern=(^$|(((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)))?
	URI        string `json:"uri,omitempty"`
	Ref        string `json:"ref,omitempty"`
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NOProxy    string `json:"noProxy,omitempty"`
	ContextDir string `json:"contextDir,omitempty"`
	SecretRef  string `json:"secretRef,omitempty"`
}

// GitOpsTrigger represents a trigge, possible type values are change, periodic, webhook.
// If token is used the object must be labeled with the following label: "gitops_config.eunomia.kohls.io/webhook_token: <token>"
type GitOpsTrigger struct {
	// Type supported types are Change, Periodic, Webhook
	// +kubebuilder:validation:Enum=Change,Periodic,Webhook
	Type string `json:"type,omitempty"`
	// creon expression only valid with the Periodic type
	Cron string `json:"cron,omitempty"`
	// webhook secret only valid with webhook type
	Secret string `json:"secret,omitempty"`
}

// GitOpsConfigSpec defines the desired state of GitOpsConfig
// +k8s:openapi-gen=true
type GitOpsConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// TemplateSource is the location of the templated resources
	TemplateSource GitConfig `json:"templateSource,omitempty"`
	// ParameterSource is the location of the parameters, only contextDir is mandatory, if other filed are left blank they are assumed to be the same as ParameterSource
	ParameterSource GitConfig `json:"parameterSource,omitempty"`
	// Triggers is an array of triggers that will lanuch this configuration
	Triggers []GitOpsTrigger `json:"triggers,omitempty"`
	// ServiceAccountRef references to the service account under which the template engine job will run, it must exists in the namespace in which this CR is created
	ServiceAccountRef string `json:"serviceAccountRef,omitempty"`
	// TemplateEngine, the gitops operator config map contains the list of available template engines, the value used here must exist in that list. Identity (i.e. no resource processing) is the default
	TemplateProcessorImage string `json:"templateProcessorImage,omitempty"`
	// ResourceHandlingMode represents how resource creation/update should be handled. Supported values are CreateOrMerge,CreateOrUpdate,Patch,None. Default is CreateOrMerge.
	// +kubebuilder:validation:Enum=CreateOrMerge,CreateOrUpdate,Patch,None
	ResourceHandlingMode string `json:"resourceHandlingMode,omitempty"`
	// ResourceDeletionMode represents how resource deletion should be handled. Supported values are Retain,Delete,None. Default is Delete
	// +kubebuilder:validation:Enum=Retain,Delete,None
	ResourceDeletionMode string `json:"resourceDeletionMode,omitempty"`
}

// GitOpsConfigStatus defines the observed state of GitOpsConfig
// +k8s:openapi-gen=true
type GitOpsConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	State            string       `json:"state,omitempty"`
	StartTime        *metav1.Time `json:"startTime,omitempty"`
	CompletionTime   *metav1.Time `json:"completionTime,omitempty"`
	Message          string       `json:"message,omitempty"`
	LastScheduleTime *metav1.Time `json:"lastScheduledTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitOpsConfig is the Schema for the gitopsconfigs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type GitOpsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsConfigSpec   `json:"spec,omitempty"`
	Status GitOpsConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitOpsConfigList contains a list of GitOpsConfig
type GitOpsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOpsConfig{}, &GitOpsConfigList{})
}
