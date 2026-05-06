# -*- mode: Python -*-
# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
# SPDX-License-Identifier: Apache-2.0

# Don"t track us.
analytics_settings(False)

update_settings(k8s_upsert_timeout_secs=60)

watch_settings(ignore=["**/*/zz_generated.deepcopy.go", 'config/crd/bases/*'])

allow_k8s_contexts(["minikube", "kind-monitoring"])

def create_temp_dir():
    from_env = os.getenv('TILT_PROMETHEUS_TEMP_DIR', '')

    if from_env != '':
        return from_env

    tmpdir = str(local("mktemp -d", echo_off=True, quiet=True)).strip()
    os.putenv('TILT_PROMETHEUS_TEMP_DIR', tmpdir)

    return tmpdir

def deploy_cert_manager():
    version = "v1.20.2"

    out = str(local("kubectl get -n cert-manager deployment/cert-manager --ignore-not-found 2>/dev/null || echo ''", quiet=True, echo_off=True))
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
    version = "v0.17.0"

    out = str(local("kubectl get -n monitoring deployment/prometheus-operator --ignore-not-found 2>/dev/null || echo ''", quiet=True, echo_off=True))
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

    print("Configuring anonymous access")
    grafana_ini = """[date_formats]
default_timezone = UTC

[auth.anonymous]
enabled = true
org_role = Admin

[auth]
disable_login_form = true
"""
    patch = '{"data":{"grafana.ini":"' + str(local("echo {} | base64".format(repr(grafana_ini)))).strip() + '"}}'
    local("kubectl patch secret grafana-config -n monitoring --type=merge -p '{}'".format(patch), quiet=True, echo_off=True)
    local("kubectl rollout restart deployment/grafana -n monitoring", quiet=True, echo_off=True)
    local("kubectl rollout status deployment/grafana -n monitoring --timeout=120s", quiet=True, echo_off=True)

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


local_resource("controller-gen", "make generate", deps=["api/", "hack/boilerplate.go.txt"])

local_resource("crds", "make install", deps=["api/"])

docker_build("controller:latest", ".", only=[
    "api/", "cmd/", "internal/", "go.mod", "go.sum"
])

deploy_cert_manager()

deploy_prometheus_operator()
setup_monitoring()

k8s_yaml(kustomize("config/develop"))
k8s_resource("monitoring-operator-controller-manager", resource_deps=["controller-gen"])

# Sample resources with manual trigger mode
k8s_yaml("./config/samples/v1alpha1_devicemonitor.yaml")
k8s_resource(new_name="DeviceMonitor", objects=["secret-basic-auth:secret", "leaf1:device", "devicemonitor-sample:devicemonitor"], trigger_mode=TRIGGER_MODE_MANUAL, auto_init=False)

print("🚀 monitoring-operator development environment")
print("👉 Edit the code inside the api/, cmd/, or internal/ directories")
print("👉 Tilt will automatically rebuild and redeploy when changes are detected")
# vim: ft=tiltfile syn=python
