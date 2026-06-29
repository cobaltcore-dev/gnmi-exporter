// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "monitoring.wire.cobaltcore.dev", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(func(s *runtime.Scheme) error {
		metav1.AddToGroupVersion(s, GroupVersion)
		return nil
	})

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// DeviceMonitorLabel is a label applied to any API object to indicate the DeviceMonitor
// that created and owns the object.
const DeviceMonitorLabel = "monitoring.wire.cobaltcore.dev/monitor-name"

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
