---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
    name: 'Monitoring Operator'
    text: 'Cloud Native Network Device Monitoring'
    tagline: 'A Kubernetes operator for automating the monitoring of network devices using gNMI telemetry'
    image:
        src: https://raw.githubusercontent.com/ironcore-dev/ironcore/refs/heads/main/docs/assets/logo_borderless.svg
        alt: Monitoring Operator
    actions:
        - theme: brand
          text: Overview
          link: /overview/
        - theme: alt
          text: API Reference
          link: /api-reference/

features:
    - title: ☸️ Kubernetes-Native Design
      details: Built with Kubebuilder and controller-runtime for seamless integration and robust operation within Kubernetes environments.
    - title: 📡 gNMI Telemetry
      details: Uses the gRPC Network Management Interface (gNMI) for real-time streaming telemetry from network devices via gnmic.
    - title: 📄 Declarative Monitoring
      details: Define monitoring targets and subscriptions as Kubernetes custom resources for a fully declarative workflow.
---
