// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the monitoring v1alpha1 API group.
// +kubebuilder:validation:Required
// +kubebuilder:object:generate=true
// +groupName=monitoring.networking.cloud.sap
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "monitoring.networking.cloud.sap", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// DeviceMonitorLabel is a label applied to any API object to indicate the DeviceMonitor
// that created and owns the object.
const DeviceMonitorLabel = "monitoring.networking.cloud.sap/monitor-name"

// Condition types that are used across different objects.
const (
	// Ready is the top-level status condition that reports if an object is ready.
	// This condition indicates whether the resource is ready to be used and will be calculated by the
	// controller based on child conditions, if present.
	ReadyCondition = "Ready"
)

// Reasons that are used across different objects.
const (
	// ReadyReason indicates that the resource is ready for use.
	ReadyReason = "Ready"

	// NotReadyReason indicates that the resource is not ready for use.
	NotReadyReason = "NotReady"

	// ReconcilePendingReason indicates that the controller is waiting for resources to be reconciled.
	ReconcilePendingReason = "ReconcilePending"
)
