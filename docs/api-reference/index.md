# API Reference

## Packages
- [monitoring.networking.cloud.sap/v1alpha1](#monitoringnetworkingcloudsapv1alpha1)


## monitoring.networking.cloud.sap/v1alpha1

Package v1alpha1 contains API Schema definitions for the monitoring.networking.cloud.sap v1alpha1 API group.

### Resource Types
- [DeviceMonitor](#devicemonitor)



#### DeviceMonitor



DeviceMonitor is the Schema for the devicemonitors API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `monitoring.networking.cloud.sap/v1alpha1` | | |
| `kind` _string_ | `DeviceMonitor` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DeviceMonitorSpec](#devicemonitorspec)_ | Specification of the desired state of the resource.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  | Required: \{\} <br /> |
| `status` _[DeviceMonitorStatus](#devicemonitorstatus)_ | Status of the resource. This is set and updated automatically.<br />Read-only.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  | Optional: \{\} <br /> |


#### DeviceMonitorSpec



DeviceMonitorSpec defines the desired state of DeviceMonitor



_Appears in:_
- [DeviceMonitor](#devicemonitor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Number of desired monitoring pods. Defaults to 1. | 1 | Minimum: 1 <br />Optional: \{\} <br /> |
| `deviceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#labelselector-v1-meta)_ | Label selector for devices that should be monitored by this DeviceMonitor. |  | Required: \{\} <br /> |
| `template` _[PodTemplateSpec](#podtemplatespec)_ | Template describes the gnmic pods that will be created. |  | Required: \{\} <br /> |
| `subscriptions` _[Subscription](#subscription) array_ | Subscriptions is a list of gNMI subscriptions to be created. |  | MinItems: 1 <br />Required: \{\} <br /> |
| `encoding` _[Encoding](#encoding)_ | Encoding represents the encoding format for gNMI messages.<br />Supported values are "json" and "json_ietf".<br />Defaults to "json_ietf". | json_ietf | Enum: [json json_ietf] <br />Optional: \{\} <br /> |
| `metrics` _[MetricsSpec](#metricsspec)_ | Metrics configures Prometheus metrics collection and ServiceMonitor<br />creation. If omitted, the metrics Service receives only<br />operator-generated labels and no ServiceMonitor is created. |  | Optional: \{\} <br /> |


#### DeviceMonitorStatus



DeviceMonitorStatus defines the observed state of DeviceMonitor.



_Appears in:_
- [DeviceMonitor](#devicemonitor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#condition-v1-meta) array_ | The conditions are a list of status objects that describe the state of the DeviceMonitor. |  | Optional: \{\} <br /> |


#### Encoding

_Underlying type:_ _string_

Encoding represents the encoding format for gNMI messages.

_Validation:_
- Enum: [json json_ietf]

_Appears in:_
- [DeviceMonitorSpec](#devicemonitorspec)

| Field | Description |
| --- | --- |
| `json` | EncodingJSON represents JSON encoding format.<br /> |
| `json_ietf` | EncodingJSONIETF represents JSON IETF encoding format.<br /> |


#### MetricsSpec



MetricsSpec configures Prometheus metrics scraping and ServiceMonitor
creation.



_Appears in:_
- [DeviceMonitorSpec](#devicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `additionalLabels` _object (keys:string, values:string)_ | AdditionalLabels are extra labels merged onto the operator-managed<br />metrics Service. They do not override operator-generated labels.<br />Use this to make the Service discoverable by an external ServiceMonitor. |  | MaxProperties: 64 <br />Optional: \{\} <br /> |
| `serviceMonitor` _[ServiceMonitorSpec](#servicemonitorspec)_ | ServiceMonitor configures the Prometheus ServiceMonitor resource.<br />If present, the operator creates and manages a ServiceMonitor.<br />If omitted, no ServiceMonitor is created. |  | Optional: \{\} <br /> |


#### Mode

_Underlying type:_ _string_

Mode represents the gNMI streaming mode.

_Validation:_
- Enum: [Stream Once Poll]

_Appears in:_
- [Subscription](#subscription)

| Field | Description |
| --- | --- |
| `Stream` | ModeStream represents the gNMI Stream mode.<br /> |
| `Once` | ModeOnce represents the gNMI Once mode.<br /> |
| `Poll` | ModePoll represents the gNMI Poll mode.<br /> |


#### PodSpec



PodSpec is a description of a pod.



_Appears in:_
- [PodTemplateSpec](#podtemplatespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |  | Optional: \{\} <br /> |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#affinity-v1-core)_ | If specified, sets the pod's scheduling constraints. |  | Optional: \{\} <br /> |


#### PodTemplateSpec



PodTemplateSpec describes the additional data a pod should have.
It is a subset from the [corev1.PodTemplateSpec] type.



_Appears in:_
- [DeviceMonitorSpec](#devicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  | Optional: \{\} <br /> |
| `spec` _[PodSpec](#podspec)_ | Specification of the desired behavior of the pod.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status |  | Optional: \{\} <br /> |


#### ServiceMonitorSpec



ServiceMonitorSpec defines configuration for the operator-managed
ServiceMonitor resource.



_Appears in:_
- [MetricsSpec](#metricsspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `additionalLabels` _object (keys:string, values:string)_ | AdditionalLabels are extra labels merged onto the ServiceMonitor<br />metadata. Use this to match the serviceMonitorSelector on a<br />Prometheus CR. |  | MaxProperties: 64 <br />Optional: \{\} <br /> |
| `interval` _[Duration](#duration)_ | Interval at which Prometheus scrapes the metrics endpoint.<br />If empty, Prometheus uses its configured global scrape interval. |  | Optional: \{\} <br /> |
| `scrapeTimeout` _[Duration](#duration)_ | ScrapeTimeout is the per-scrape timeout when querying the metrics<br />endpoint. Must be less than or equal to Interval.<br />If empty, Prometheus uses its configured global scrape timeout. |  | Optional: \{\} <br /> |


#### StreamMode

_Underlying type:_ _string_

StreamMode represents the gNMI stream mode.

_Validation:_
- Enum: [TargetDefined Sample OnChange]

_Appears in:_
- [Subscription](#subscription)

| Field | Description |
| --- | --- |
| `TargetDefined` | StreamModeTargetDefined represents the gNMI TargetDefined stream mode.<br /> |
| `Sample` | StreamModeSample represents the gNMI Sample stream mode.<br /> |
| `OnChange` | StreamModeOnChange represents the gNMI OnChange stream mode.<br /> |


#### Subscription







_Appears in:_
- [DeviceMonitorSpec](#devicemonitorspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the subscription. |  | MaxLength: 255 <br />MinLength: 1 <br />Required: \{\} <br /> |
| `paths` _string array_ | Paths is a list of gNMI paths to subscribe to. |  | MaxItems: 10 <br />MinItems: 1 <br />items:MaxLength: 1024 <br />items:MinLength: 1 <br />Required: \{\} <br /> |
| `models` _string array_ | Models is a list of YANG models to be used for the subscription. |  | MaxItems: 10 <br />MinItems: 0 <br />items:MaxLength: 255 <br />items:MinLength: 1 <br />Optional: \{\} <br /> |
| `mode` _[Mode](#mode)_ | Mode represents the gNMI streaming mode.<br />Supported values are "Stream", "Once", and "Poll".<br />Defaults to "Stream". | Stream | Enum: [Stream Once Poll] <br />Optional: \{\} <br /> |
| `streamMode` _[StreamMode](#streammode)_ | StreamMode represents the gNMI stream mode.<br />Supported values are "TargetDefined", "Sample", and "OnChange".<br />This field is only applicable when Mode is set to "Stream".<br />Defaults to "Sample". | Sample | Enum: [TargetDefined Sample OnChange] <br />Optional: \{\} <br /> |
| `sampleInterval` _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#duration-v1-meta)_ | SampleInterval is the interval at which to sample data.<br />This field is only applicable when StreamMode is set to "Sample".<br />The value must be a valid duration string, e.g., "10s", "1m", "1h".<br />Defaults to "10s". | 10s | Format: duration <br />Pattern: `^([0-9]+(ns\|us\|ms\|s\|m\|h))+$` <br />Type: string <br />Optional: \{\} <br /> |


