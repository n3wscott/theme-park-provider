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
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RideParameters struct {
	// Type of Ride.
	Type string `json:"type"`
	// Capacity is the riders per trip supported on this ride.
	Capacity int `json:"capacity"`
}

// RideSpec defines the desired state of Ride.
type RideSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	ForProvider RideParameters `json:"forProvider"`
}

// RideStatus defines the observed state of Ride.
type RideStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// Operator is the operator assigned to this Ride.
	// +optional
	Operator *xpv1.TypedReference `json:"operator"`

	RidersPerHour int `json:"ridersPerHour"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Ride is the Schema for the rides API.
type Ride struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RideSpec   `json:"spec,omitempty"`
	Status RideStatus `json:"status,omitempty"`
}

var _ resource.Managed = (*Ride)(nil)

// +kubebuilder:object:root=true

// RideList contains a list of Ride.
type RideList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ride `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ride{}, &RideList{})
}
