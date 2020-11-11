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

package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/dsyer/spring-config-operator/api/v1"
)

// +kubebuilder:rbac:groups=spring.io,resources=configclients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=spring.io,resources=configclients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

var (
	ownerKey = ".metadata.controller"
	apiGVStr = api.GroupVersion.String()
)

// ConfigClientReconciler reconciles a ConfigClient object
func ConfigClientReconciler(c reconcilers.Config) *reconcilers.ParentReconciler {
	c.Log = c.Log.WithName("ConfigClient")

	return &reconcilers.ParentReconciler{
		Type: &api.ConfigClient{},
		SubReconcilers: []reconcilers.SubReconciler{
			ConfigMapReconciler(c),
		},

		Config: c,
	}
}

// ConfigMapReconciler creates a new ConfigMap if needed
func ConfigMapReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	c.Log = c.Log.WithName("ConfigMap")
	return &reconcilers.ChildReconciler{
		Config:        c,
		ParentType:    &api.ConfigClient{},
		ChildType:     &corev1.ConfigMap{},
		ChildListType: &corev1.ConfigMapList{},

		DesiredChild: func(client *api.ConfigClient) (*corev1.ConfigMap, error) {
			return extract(client, c), nil
		},

		ReflectChildStatusOnParent: func(client *api.ConfigClient, child *corev1.ConfigMap, err error) {
		},

		MergeBeforeUpdate: func(current, desired *corev1.ConfigMap) {
			current.Labels = desired.Labels
		},

		SemanticEquals: func(a1, a2 *corev1.ConfigMap) bool {
			return equality.Semantic.DeepEqual(a1.Labels, a2.Labels) && equality.Semantic.DeepEqual(a1.Data, a2.Data)
		},

		IndexField: ".metadata.configClientConfigMap",

		Sanitize: func(child *corev1.ConfigMap) interface{} {
			return child.Data
		},
	}
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// PropertySource respresents a named set of name-value pairs
type PropertySource struct {
	Name   string            `json:"name"`
	Source map[string]string `json:"source"`
}

// Environment represents the result from a confg server
type Environment struct {
	PropertySources []PropertySource `json:"propertySources"`
}

func getJSON(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func extract(client *api.ConfigClient, c reconcilers.Config) *corev1.ConfigMap {
	config := corev1.ConfigMap{}
	config.Name = client.Name
	config.Namespace = client.Namespace
	config.Labels = client.Labels
	environment := &Environment{}
	config.Data = map[string]string{}
	if err := getJSON(client.Spec.URL, environment); err != nil {
		c.Log.Error(err, "Failed to download config data", "configclient", client.Name, "url", client.Spec.URL)
		client.Status.Complete = false
		return nil
	}
	for _, properties := range environment.PropertySources {
		for k, v := range properties.Source {
			if _, ok := config.Data[k]; !ok {
				config.Data[k] = v
			}
		}
	}
	c.Log.Info("Downloaded config data", "configclient", client.Name, "url", client.Spec.URL)
	client.Status.Complete = true
	return &config
}
