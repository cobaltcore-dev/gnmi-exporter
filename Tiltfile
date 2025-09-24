# -*- mode: Python -*-
# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
# SPDX-License-Identifier: Apache-2.0

# Don"t track us.
analytics_settings(False)

update_settings(k8s_upsert_timeout_secs=60)

allow_k8s_contexts(["kind-monitoring"])

def create_temp_dir():
    from_env = os.getenv('TILT_PROMETHEUS_TEMP_DIR', '')

    if from_env != '':
        return from_env

    tmpdir = str(local("mktemp -d", echo_off=True, quiet=True)).strip()
    os.putenv('TILT_PROMETHEUS_TEMP_DIR', tmpdir)

    return tmpdir

def deploy_cert_manager():
    version = "v1.18.2"

    out = str(local("kubectl get -n cert-manager deployment/cert-manager 2>/dev/null || echo ''", quiet=True, echo_off=True))
    if out == version:
        print("cert-manager already installed")
        return

    print("Installing cert-manager")
    local("kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/{}/cert-manager.yaml".format(version), quiet=True, echo_off=True)

    print("Waiting for cert-manager to start")
    local("kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager", quiet=True, echo_off=True)
    local("kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-cainjector", quiet=True, echo_off=True)
    local("kubectl wait --for=condition=Available --timeout=300s -n cert-manager deployment/cert-manager-webhook", quiet=True, echo_off=True)

def deploy_prometheus_operator():
    version = "v0.16.0"

    out = str(local("kubectl get -n monitoring deployment/prometheus-operator 2>/dev/null || echo ''", quiet=True, echo_off=True))
    if out != "":
        print("kube-prometheus stack already installed")
        return

    temp_dir = create_temp_dir()

    print("Installing kube-prometheus stack")
    local("curl -fsSL https://github.com/prometheus-operator/kube-prometheus/archive/refs/tags/{}.tar.gz | tar xzf - --strip-components=1 -C {}".format(version, temp_dir), quiet=True, echo_off=True)

    local("kubectl apply --server-side -f {}/manifests/setup".format(temp_dir), quiet=True, echo_off=True)
    local("kubectl wait --for=condition=Established --timeout=300s crd/servicemonitors.monitoring.coreos.com", quiet=True, echo_off=True)
    local("kubectl apply --server-side --force-conflicts -f {}/manifests".format(temp_dir), quiet=True, echo_off=True)

    print("Waiting for grafana to start")
    local("kubectl wait --for=condition=Available --timeout=300s -n monitoring deployment/grafana", quiet=True, echo_off=True)

    print("Waiting for prometheus-operator to start")
    local("kubectl wait --for=condition=Available --timeout=300s -n monitoring deployment/prometheus-operator", quiet=True, echo_off=True)

def setup_monitoring():
    resources = {
        "grafana": {
            "port_forward": "3000",
            "service": "grafana",
        },
        "prometheus": {
            "port_forward": "9090",
            "service": "prometheus-operated",
        },
        "alertmanager": {
            "port_forward": "9093",
            "service": "alertmanager-operated",
        },
    }

    for name, resource in resources.items():
        service = resource.get("service")
        port = resource.get("port_forward")
        url = "http://localhost:{}".format(port.split(":")[0])
        cmd = "kubectl wait --for=condition=Ready --timeout=300s -n monitoring pod -l app.kubernetes.io/name={}; kubectl port-forward -n monitoring svc/{} {}".format(name, service, port)
        local_resource(name, serve_cmd=cmd, links=[link(url, name)])


docker_build("controller:latest", ".", ssh='default', ignore=["*/*/zz_generated.deepcopy.go", "config/crd/bases/*"], only=[
    "api/", "cmd/", "hack/", "internal/", "go.mod", "go.sum", "Makefile",
])

local_resource("controller-gen", "make generate", ignore=["*/*/zz_generated.deepcopy.go", "config/crd/bases/*"], deps=[
    "api/", "cmd/", "hack/", "internal/", "go.mod", "go.sum", "Makefile",
])

deploy_cert_manager()

deploy_prometheus_operator()
setup_monitoring()

k8s_yaml(kustomize("config/develop"))
k8s_resource("monitoring-operator-controller-manager", resource_deps=["controller-gen"])

# Sample resources with manual trigger mode
k8s_yaml("./config/samples/monitoring_v1alpha1_devicemonitor.yaml")
k8s_resource(new_name="DeviceMonitor", objects=["secret-basic-auth:secret", "leaf1:device", "devicemonitor-sample:devicemonitor"], trigger_mode=TRIGGER_MODE_MANUAL, auto_init=False)

print("🚀 monitoring-operator development environment")
print("👉 Edit the code inside the api/, cmd/, or internal/ directories")
print("👉 Tilt will automatically rebuild and redeploy when changes are detected")
# vim: ft=tiltfile syn=python
