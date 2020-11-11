/*

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

// ConfigClientSpec defines the desired state of ConfigClient
type ConfigClientSpec struct {
	URL string `json:"url,omitempty"`
}

// ConfigClientStatus defines the observed state of ConfigClient
type ConfigClientStatus struct {
	Complete           bool  `json:"complete,omitempty"`
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigClient is the Schema for the springs API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".spec.url",description="config url"
// +kubebuilder:printcolumn:name="Complete",type="boolean",JSONPath=".status.complete",description="config status"
type ConfigClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigClientSpec   `json:"spec,omitempty"`
	Status ConfigClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigClientList contains a list of Spring
type ConfigClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigClient{}, &ConfigClientList{})
}
