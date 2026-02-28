// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=usage.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This isn't a custom resource, in the sense that we never install its CRD.
// It is a KRM-like object, so we generate a CRD to describe its schema.

// Input can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Usages defines the usages between composed resources
	Usages []Usage `json:"usages"`
}

// Usage is a simplified view of a Crossplane Usage used in compositions..
type Usage struct {
	By Resource `json:"by"`
	Of Resource `json:"of"`
}

// Resource is the specification of composed resource(s).
type Resource struct {
	Name string `json:"name"`
}
