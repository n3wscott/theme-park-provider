/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RideOperatorParameters struct {
	// Frequency is how often this operator operates ride per hour.
	Frequency int `json:"frequency"`

	// Ride is the ride this operator is assigned to.
	// +optional
	Ride *xpv1.TypedReference `json:"ride"`
}

// RideOperatorSpec defines the desired state of RideOperator.
type RideOperatorSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	ForProvider RideOperatorParameters `json:"forProvider"`
}

// RideOperatorStatus defines the observed state of RideOperator.
type RideOperatorStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// RideOperator is the Schema for the rideoperators API.
type RideOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RideOperatorSpec   `json:"spec,omitempty"`
	Status RideOperatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RideOperatorList contains a list of RideOperator.
type RideOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RideOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RideOperator{}, &RideOperatorList{})
}
