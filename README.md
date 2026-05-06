<!--
# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
# SPDX-License-Identifier: Apache-2.0
-->

# monitoring-operator

[![REUSE status](https://api.reuse.software/badge/github.com/cobaltcore-dev/monitoring-operator)](https://api.reuse.software/info/github.com/cobaltcore-dev/monitoring-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/cobaltcore-dev/monitoring-operator)](https://goreportcard.com/report/github.com/cobaltcore-dev/monitoring-operator)
[![GitHub License](https://img.shields.io/static/v1?label=License&message=Apache-2.0&color=blue)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)

`monitoring-operator` is a Kubernetes operator for automating the monitoring of network devices.

## Description

Monitoring-operator is a project built using Kubebuilder and controller-runtime to facilitate the monitoring of network devices managed via [network-operator](https://github.com/ironcore-dev/network-operator). It provides a robust and scalable solution for monitoring networking infrastructure, ensuring seamless integration and automation within Kubernetes environments.

It's highly recommended to watch the following Talk on [Improving Network Observability with Telemetry Using gNMIc and Prometheus](https://www.youtube.com/watch?v=2X2F_L622g8) to familiarize some of the concepts used in this project.

## Getting Started

### Prerequisites

- go version v1.26.0+
- docker version 28+.
- kubectl version v1.33.1+.
- Access to a Kubernetes v1.33.0+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/monitoring-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. And it is required to have access to pull the image from the working environment. Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make deploy-crds
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/monitoring-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples have default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make undeploy-crds
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/monitoring-operator:tag
```

> NOTE: The makefile target mentioned above generates an 'install.yaml' file in the dist directory. This file contains all the resources built with Kustomize, which are necessary to install this project without its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/monitoring-operator/<tag or branch>/dist/install.yaml
```

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/cobaltcore-dev/monitoring-operator/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure

If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/cobaltcore-dev/monitoring-operator/security/policy) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and CobaltCore contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/cobaltcore-dev/monitoring-operator).
