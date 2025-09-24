// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceMonitorSpec defines the desired state of DeviceMonitor
type DeviceMonitorSpec struct {
	// Number of desired monitoring pods. Defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Label selector for devices that should be monitored by this DeviceMonitor.
	// +required
	Selector metav1.LabelSelector `json:"deviceSelector"`

	// Template describes the gnmic pods that will be created.
	// +required
	Template PodTemplateSpec `json:"template"`

	// Subscriptions is a list of gNMI subscriptions to be created.
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=name
	// +required
	Subscriptions []Subscription `json:"subscriptions"`

	// Encoding represents the encoding format for gNMI messages.
	// Supported values are "json" and "json_ietf".
	// Defaults to "json_ietf".
	// +kubebuilder:default="json_ietf"
	// +optional
	Encoding Encoding `json:"encoding,omitempty"`
}

// PodTemplateSpec describes the additional data a pod should have.
// It is a subset from the [corev1.PodTemplateSpec] type.
type PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec PodSpec `json:"spec,omitempty,omitzero"`
}

// PodSpec is a description of a pod.
type PodSpec struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, sets the pod's scheduling constraints.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

type Subscription struct {
	// Name is the name of the subscription.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +required
	Name string `json:"name"`

	// Paths is a list of gNMI paths to subscribe to.
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=1024
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +listType=set
	// +required
	Paths []string `json:"paths,omitempty"`

	// Models is a list of YANG models to be used for the subscription.
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=255
	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=10
	// +listType=set
	// +optional
	Models []string `json:"models,omitempty"`

	// Mode represents the gNMI streaming mode.
	// Supported values are "Stream", "Once", and "Poll".
	// Defaults to "Stream".
	// +kubebuilder:default="Stream"
	// +optional
	Mode Mode `json:"mode,omitempty"`

	// StreamMode represents the gNMI stream mode.
	// Supported values are "TargetDefined", "Sample", and "OnChange".
	// This field is only applicable when Mode is set to "Stream".
	// Defaults to "Sample".
	// +kubebuilder:default="Sample"
	// +optional
	StreamMode StreamMode `json:"streamMode,omitempty"`

	// SampleInterval is the interval at which to sample data.
	// This field is only applicable when StreamMode is set to "Sample".
	// The value must be a valid duration string, e.g., "10s", "1m", "1h".
	// Defaults to "10s".
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(ns|us|ms|s|m|h))+$`
	// +kubebuilder:default="10s"
	// +optional
	SampleInterval metav1.Duration `json:"sampleInterval,omitempty,omitzero"`
}

// Mode represents the gNMI streaming mode.
// +kubebuilder:validation:Enum=Stream;Once;Poll
type Mode string

const (
	// ModeStream represents the gNMI Stream mode.
	ModeStream Mode = "Stream"
	// ModeOnce represents the gNMI Once mode.
	ModeOnce Mode = "Once"
	// ModePoll represents the gNMI Poll mode.
	ModePoll Mode = "Poll"
)

// Config returns the gnmic configuration string representation of the Mode.
func (m Mode) Config() string {
	switch m {
	case ModeStream:
		return "STREAM"
	case ModeOnce:
		return "ONCE"
	case ModePoll:
		return "POLL"
	default:
		return ""
	}
}

// StreamMode represents the gNMI stream mode.
// +kubebuilder:validation:Enum=TargetDefined;Sample;OnChange
type StreamMode string

const (
	// StreamModeTargetDefined represents the gNMI TargetDefined stream mode.
	StreamModeTargetDefined StreamMode = "TargetDefined"
	// StreamModeSample represents the gNMI Sample stream mode.
	StreamModeSample StreamMode = "Sample"
	// StreamModeOnChange represents the gNMI OnChange stream mode.
	StreamModeOnChange StreamMode = "OnChange"
)

// Config returns the gnmic configuration string representation of the StreamMode.
func (sm StreamMode) Config() string {
	switch sm {
	case StreamModeTargetDefined:
		return "TARGET_DEFINED"
	case StreamModeSample:
		return "SAMPLE"
	case StreamModeOnChange:
		return "ON_CHANGE"
	default:
		return ""
	}
}

// Encoding represents the encoding format for gNMI messages.
// +kubebuilder:validation:Enum=json;json_ietf
type Encoding string

const (
	// EncodingJSON represents JSON encoding format.
	EncodingJSON Encoding = "json"
	// EncodingJSONIETF represents JSON IETF encoding format.
	EncodingJSONIETF Encoding = "json_ietf"
)

// DeviceMonitorStatus defines the observed state of DeviceMonitor.
type DeviceMonitorStatus struct {
	// The conditions are a list of status objects that describe the state of the DeviceMonitor.
	//+listType=map
	//+listMapKey=type
	//+patchStrategy=merge
	//+patchMergeKey=type
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=devicemonitors
// +kubebuilder:resource:singular=devicemonitor
// +kubebuilder:resource:shortName=dm
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// DeviceMonitor is the Schema for the devicemonitors API
type DeviceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// Specification of the desired state of the resource.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +required
	Spec DeviceMonitorSpec `json:"spec"`

	// Status of the resource. This is set and updated automatically.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status DeviceMonitorStatus `json:"status,omitempty,omitzero"`
}

func (m *DeviceMonitor) ServiceName() string {
	// Service name must match "${cluster-name}-gnmic-api" for gnmic clustering to work
	// See: https://github.com/openconfig/gnmic/blob/0aa04b5727cd894ef7ef9e8f787dc2f9c513e5b0/pkg/app/api.go#L318
	return m.Name + "-gnmic-api"
}

func (m *DeviceMonitor) SetReadyCondition(status metav1.ConditionStatus, reason, message string) {
	cond := metav1.Condition{
		Type:               ReadyCondition,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: m.Generation,
	}

	meta.SetStatusCondition(&m.Status.Conditions, cond)
}

// +kubebuilder:object:root=true

// DeviceMonitorList contains a list of DeviceMonitor
type DeviceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty,omitzero"`
	Items           []DeviceMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceMonitor{}, &DeviceMonitorList{})
}
