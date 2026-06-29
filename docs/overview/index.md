# Introduction

`gnmi-exporter` is a Kubernetes operator for automating the monitoring of network devices. It deploys and manages [gnmic](https://gnmic.openconfig.net/) instances that collect gNMI streaming telemetry from network devices managed by [network-operator](https://github.com/ironcore-dev/network-operator).

## How it works

1. You create a `DeviceMonitor` custom resource specifying which devices to monitor and what gNMI subscriptions to configure.
2. The operator deploys gnmic pods that connect to the selected devices and stream telemetry data.
3. The collected metrics are exposed in Prometheus format for integration with your observability stack.

## Prerequisites

- Kubernetes v1.33.0+
- [network-operator](https://github.com/ironcore-dev/network-operator) managing the target devices
