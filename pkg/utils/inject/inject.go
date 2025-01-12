/*
 Copyright 2023 The KusionStack Authors.

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

package inject

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "kusionstack.io/operating/apis/apps/v1alpha1"
)

const (
	FieldIndexOwnerRefUID       = "ownerRefUID"
	FieldIndexPodTransitionRule = "podtransitionruleIndex"
)

func NewCacheWithFieldIndex(config *rest.Config, opts cache.Options) (cache.Cache, error) {
	c, err := cache.New(config, opts)
	if err != nil {
		return c, err
	}

	c.IndexField(context.TODO(), &corev1.Pod{}, FieldIndexOwnerRefUID, func(pod client.Object) []string {
		ownerRef := metav1.GetControllerOf(pod)
		if ownerRef == nil {
			return nil
		}

		return []string{string(ownerRef.UID)}
	})

	c.IndexField(context.TODO(), &appv1.ControllerRevision{}, FieldIndexOwnerRefUID, func(revision client.Object) []string {
		ownerRef := metav1.GetControllerOf(revision)
		if ownerRef == nil {
			return nil
		}

		return []string{string(ownerRef.UID)}
	})

	c.IndexField(context.TODO(), &appsv1alpha1.PodTransitionRule{}, FieldIndexPodTransitionRule, func(obj client.Object) []string {
		rs, ok := obj.(*appsv1alpha1.PodTransitionRule)
		if !ok {
			return nil
		}
		return rs.Status.Targets
	})
	return c, err
}
