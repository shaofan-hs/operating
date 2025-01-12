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

package collaset

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	k8scorev1 "k8s.io/kubernetes/pkg/apis/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1alpha1 "kusionstack.io/operating/apis/apps/v1alpha1"
	commonutils "kusionstack.io/operating/pkg/utils"
	"kusionstack.io/operating/pkg/utils/mixin"
)

type MutatingHandler struct {
	*mixin.WebhookHandlerMixin
}

func NewMutatingHandler() *MutatingHandler {
	return &MutatingHandler{
		WebhookHandlerMixin: mixin.NewWebhookHandlerMixin(),
	}
}

func (h *MutatingHandler) Handle(ctx context.Context, req admission.Request) (resp admission.Response) {
	if req.Operation != admissionv1.Update && req.Operation != admissionv1.Create {
		return admission.Allowed("")
	}

	logger := h.Logger.WithValues(
		"op", req.Operation,
		"collaset", commonutils.AdmissionRequestObjectKeyString(req),
	)

	cls := &appsv1alpha1.CollaSet{}
	if err := h.Decoder.Decode(req, cls); err != nil {
		logger.Error(err, "failed to decode collaset")
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.setDetaultCollaSet(cls)
	marshalled, err := json.Marshal(cls)
	if err != nil {
		logger.Error(err, "failed to marshal collaset to json")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshalled)
}

func (h *MutatingHandler) setDetaultCollaSet(cls *appsv1alpha1.CollaSet) {
	h.setDefaultPodSpec(cls)
	h.setDefaultCollaSetUpdateStrategy(cls)
}

func (h *MutatingHandler) setDefaultPodSpec(in *appsv1alpha1.CollaSet) {
	k8scorev1.SetDefaults_PodSpec(&in.Spec.Template.Spec)
	for i := range in.Spec.Template.Spec.Volumes {
		a := &in.Spec.Template.Spec.Volumes[i]
		k8scorev1.SetDefaults_Volume(a)
		if a.VolumeSource.HostPath != nil {
			k8scorev1.SetDefaults_HostPathVolumeSource(a.VolumeSource.HostPath)
		}
		if a.VolumeSource.Secret != nil {
			k8scorev1.SetDefaults_SecretVolumeSource(a.VolumeSource.Secret)
		}
		if a.VolumeSource.ISCSI != nil {
			k8scorev1.SetDefaults_ISCSIVolumeSource(a.VolumeSource.ISCSI)
		}
		if a.VolumeSource.RBD != nil {
			k8scorev1.SetDefaults_RBDVolumeSource(a.VolumeSource.RBD)
		}
		if a.VolumeSource.DownwardAPI != nil {
			k8scorev1.SetDefaults_DownwardAPIVolumeSource(a.VolumeSource.DownwardAPI)
			for j := range a.VolumeSource.DownwardAPI.Items {
				b := &a.VolumeSource.DownwardAPI.Items[j]
				if b.FieldRef != nil {
					k8scorev1.SetDefaults_ObjectFieldSelector(b.FieldRef)
				}
			}
		}
		if a.VolumeSource.ConfigMap != nil {
			k8scorev1.SetDefaults_ConfigMapVolumeSource(a.VolumeSource.ConfigMap)
		}
		if a.VolumeSource.AzureDisk != nil {
			k8scorev1.SetDefaults_AzureDiskVolumeSource(a.VolumeSource.AzureDisk)
		}
		if a.VolumeSource.Projected != nil {
			k8scorev1.SetDefaults_ProjectedVolumeSource(a.VolumeSource.Projected)
			for j := range a.VolumeSource.Projected.Sources {
				b := &a.VolumeSource.Projected.Sources[j]
				if b.DownwardAPI != nil {
					for k := range b.DownwardAPI.Items {
						c := &b.DownwardAPI.Items[k]
						if c.FieldRef != nil {
							k8scorev1.SetDefaults_ObjectFieldSelector(c.FieldRef)
						}
					}
				}
				if b.ServiceAccountToken != nil {
					k8scorev1.SetDefaults_ServiceAccountTokenProjection(b.ServiceAccountToken)
				}
			}
		}
		if a.VolumeSource.ScaleIO != nil {
			k8scorev1.SetDefaults_ScaleIOVolumeSource(a.VolumeSource.ScaleIO)
		}
		if a.VolumeSource.Ephemeral != nil {
			if a.VolumeSource.Ephemeral.VolumeClaimTemplate != nil {
				k8scorev1.SetDefaults_PersistentVolumeClaimSpec(&a.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec)
				k8scorev1.SetDefaults_ResourceList(&a.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.Resources.Limits)
				k8scorev1.SetDefaults_ResourceList(&a.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.Resources.Requests)
			}
		}
	}
	for i := range in.Spec.Template.Spec.InitContainers {
		a := &in.Spec.Template.Spec.InitContainers[i]
		k8scorev1.SetDefaults_Container(a)
		for j := range a.Ports {
			b := &a.Ports[j]
			if b.Protocol == "" {
				b.Protocol = "TCP"
			}
		}
		for j := range a.Env {
			b := &a.Env[j]
			if b.ValueFrom != nil {
				if b.ValueFrom.FieldRef != nil {
					k8scorev1.SetDefaults_ObjectFieldSelector(b.ValueFrom.FieldRef)
				}
			}
		}
		k8scorev1.SetDefaults_ResourceList(&a.Resources.Limits)
		k8scorev1.SetDefaults_ResourceList(&a.Resources.Requests)
		if a.LivenessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.LivenessProbe)
			if a.LivenessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.LivenessProbe.Handler.HTTPGet)
			}
		}
		if a.ReadinessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.ReadinessProbe)
			if a.ReadinessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.ReadinessProbe.Handler.HTTPGet)
			}
		}
		if a.StartupProbe != nil {
			k8scorev1.SetDefaults_Probe(a.StartupProbe)
			if a.StartupProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.StartupProbe.Handler.HTTPGet)
			}
		}
		if a.Lifecycle != nil {
			if a.Lifecycle.PostStart != nil {
				if a.Lifecycle.PostStart.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.Lifecycle.PostStart.HTTPGet)
				}
			}
			if a.Lifecycle.PreStop != nil {
				if a.Lifecycle.PreStop.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.Lifecycle.PreStop.HTTPGet)
				}
			}
		}
	}
	for i := range in.Spec.Template.Spec.Containers {
		a := &in.Spec.Template.Spec.Containers[i]
		k8scorev1.SetDefaults_Container(a)
		for j := range a.Ports {
			b := &a.Ports[j]
			if b.Protocol == "" {
				b.Protocol = "TCP"
			}
		}
		for j := range a.Env {
			b := &a.Env[j]
			if b.ValueFrom != nil {
				if b.ValueFrom.FieldRef != nil {
					k8scorev1.SetDefaults_ObjectFieldSelector(b.ValueFrom.FieldRef)
				}
			}
		}
		k8scorev1.SetDefaults_ResourceList(&a.Resources.Limits)
		k8scorev1.SetDefaults_ResourceList(&a.Resources.Requests)
		if a.LivenessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.LivenessProbe)
			if a.LivenessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.LivenessProbe.Handler.HTTPGet)
			}
		}
		if a.ReadinessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.ReadinessProbe)
			if a.ReadinessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.ReadinessProbe.Handler.HTTPGet)
			}
		}
		if a.StartupProbe != nil {
			k8scorev1.SetDefaults_Probe(a.StartupProbe)
			if a.StartupProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.StartupProbe.Handler.HTTPGet)
			}
		}
		if a.Lifecycle != nil {
			if a.Lifecycle.PostStart != nil {
				if a.Lifecycle.PostStart.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.Lifecycle.PostStart.HTTPGet)
				}
			}
			if a.Lifecycle.PreStop != nil {
				if a.Lifecycle.PreStop.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.Lifecycle.PreStop.HTTPGet)
				}
			}
		}
	}
	for i := range in.Spec.Template.Spec.EphemeralContainers {
		a := &in.Spec.Template.Spec.EphemeralContainers[i]
		k8scorev1.SetDefaults_EphemeralContainer(a)
		for j := range a.EphemeralContainerCommon.Ports {
			b := &a.EphemeralContainerCommon.Ports[j]
			if b.Protocol == "" {
				b.Protocol = "TCP"
			}
		}
		for j := range a.EphemeralContainerCommon.Env {
			b := &a.EphemeralContainerCommon.Env[j]
			if b.ValueFrom != nil {
				if b.ValueFrom.FieldRef != nil {
					k8scorev1.SetDefaults_ObjectFieldSelector(b.ValueFrom.FieldRef)
				}
			}
		}
		k8scorev1.SetDefaults_ResourceList(&a.EphemeralContainerCommon.Resources.Limits)
		k8scorev1.SetDefaults_ResourceList(&a.EphemeralContainerCommon.Resources.Requests)
		if a.EphemeralContainerCommon.LivenessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.EphemeralContainerCommon.LivenessProbe)
			if a.EphemeralContainerCommon.LivenessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.EphemeralContainerCommon.LivenessProbe.Handler.HTTPGet)
			}
		}
		if a.EphemeralContainerCommon.ReadinessProbe != nil {
			k8scorev1.SetDefaults_Probe(a.EphemeralContainerCommon.ReadinessProbe)
			if a.EphemeralContainerCommon.ReadinessProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.EphemeralContainerCommon.ReadinessProbe.Handler.HTTPGet)
			}
		}
		if a.EphemeralContainerCommon.StartupProbe != nil {
			k8scorev1.SetDefaults_Probe(a.EphemeralContainerCommon.StartupProbe)
			if a.EphemeralContainerCommon.StartupProbe.Handler.HTTPGet != nil {
				k8scorev1.SetDefaults_HTTPGetAction(a.EphemeralContainerCommon.StartupProbe.Handler.HTTPGet)
			}
		}
		if a.EphemeralContainerCommon.Lifecycle != nil {
			if a.EphemeralContainerCommon.Lifecycle.PostStart != nil {
				if a.EphemeralContainerCommon.Lifecycle.PostStart.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.EphemeralContainerCommon.Lifecycle.PostStart.HTTPGet)
				}
			}
			if a.EphemeralContainerCommon.Lifecycle.PreStop != nil {
				if a.EphemeralContainerCommon.Lifecycle.PreStop.HTTPGet != nil {
					k8scorev1.SetDefaults_HTTPGetAction(a.EphemeralContainerCommon.Lifecycle.PreStop.HTTPGet)
				}
			}
		}
	}
	k8scorev1.SetDefaults_ResourceList(&in.Spec.Template.Spec.Overhead)
}

func (h *MutatingHandler) setDefaultCollaSetUpdateStrategy(cls *appsv1alpha1.CollaSet) {
	if cls.Spec.UpdateStrategy.PodUpdatePolicy == "" {
		cls.Spec.UpdateStrategy.PodUpdatePolicy = appsv1alpha1.CollaSetInPlaceIfPossiblePodUpdateStrategyType
	}

	if cls.Spec.UpdateStrategy.RollingUpdate == nil {
		cls.Spec.UpdateStrategy.RollingUpdate = &appsv1alpha1.RollingUpdateCollaSetStrategy{}
	}

	if cls.Spec.UpdateStrategy.RollingUpdate.ByPartition == nil && cls.Spec.UpdateStrategy.RollingUpdate.ByLabel == nil {
		cls.Spec.UpdateStrategy.RollingUpdate.ByPartition = &appsv1alpha1.ByPartition{}
	}
}

var _ inject.Client = &MutatingHandler{}

func (h *MutatingHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}

var _ admission.DecoderInjector = &MutatingHandler{}

func (h *MutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
