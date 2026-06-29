# Quickstart

Get the gNMI exporter running in your cluster and collecting network telemetry in minutes.

## Installation

Install the monitoring operator using the Helm chart published as an OCI artifact.

Available versions can be found at the [GitHub Container Registry](https://github.com/cobaltcore-dev/gnmi-exporter/pkgs/container/charts%2Fgnmi-exporter).

```bash
helm install gnmi-exporter \
  oci://ghcr.io/cobaltcore-dev/charts/gnmi-exporter \
  --version 0.0.0-36c6d53 \
  --namespace monitoring-system \
  --create-namespace
```

## Creating a DeviceMonitor

A `DeviceMonitor` selects network devices by label and configures gNMI subscriptions for telemetry collection:

```yaml
apiVersion: monitoring.networking.cloud.sap/v1alpha1
kind: DeviceMonitor
metadata:
  name: spine-monitor
spec:
  replicas: 2
  deviceSelector:
    matchLabels:
      networking.metal.ironcore.dev/device-role: evpn-spine
  template:
    spec:
      nodeSelector:
        topology.kubernetes.io/zone: eu-de-1a
  subscriptions:
    - name: interface-counters
      paths:
        - /interfaces/interface/state/counters
      mode: Stream
      streamMode: Sample
      sampleInterval: 10s
  encoding: json_ietf
```

## Prometheus Integration

The operator always creates a headless metrics Service exposing port `9804`. You can integrate with Prometheus in two ways.

### Option A: Use an existing ServiceMonitor

If you already have a ServiceMonitor (or a Prometheus scrape configuration) that discovers Services by label, add custom labels to the metrics Service using `spec.metrics.additionalLabels`:

```yaml
spec:
  metrics:
    additionalLabels:
      monitoring.example.com/scrape: "true"
```

The operator merges these labels onto the created Service. Your existing ServiceMonitor can then select it:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: networking-services
spec:
  selector:
    matchLabels:
      monitoring.example.com/scrape: "true"
  endpoints:
    - port: http
      path: /metrics
```

### Option B: Let the operator create a ServiceMonitor

If you want the operator to manage its own ServiceMonitor, set `spec.metrics.serviceMonitor`:

```yaml
spec:
  metrics:
    serviceMonitor:
      additionalLabels:
        prometheus: infrastructure
      interval: "30s"
      scrapeTimeout: "10s"
```

The operator creates a ServiceMonitor targeting the metrics Service. Use `additionalLabels` to match the `serviceMonitorSelector` on your Prometheus CR so it gets picked up automatically.

### No monitoring resources (default)

When `spec.metrics` is omitted entirely, the operator does not create a ServiceMonitor. The metrics Service still exists and can be scraped manually or via pod annotations.
