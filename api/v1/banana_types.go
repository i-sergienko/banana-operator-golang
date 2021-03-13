/*
Copyright 2021.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BananaSpec defines the desired state of Banana
type BananaSpec struct {
	Color string `json:"color,omitempty"`
}

// BananaStatus defines the observed state of Banana
type BananaStatus struct {
	Color string `json:"color,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Banana is the Schema for the bananas API
type Banana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BananaSpec   `json:"spec,omitempty"`
	Status BananaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BananaList contains a list of Banana
type BananaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Banana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Banana{}, &BananaList{})
}
