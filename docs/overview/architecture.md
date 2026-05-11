<script setup>
import { useData } from 'vitepress'
import { computed } from 'vue'
import architectureLight from '../assets/gnmi-monitoring.svg?url'
import architectureDark from '../assets/gnmi-monitoring-dark.svg?url'

const { isDark } = useData()
const architectureImage = computed(() =>
  isDark.value ? architectureDark : architectureLight
)
</script>

# Architecture

The monitoring-operator automates gNMI streaming telemetry collection from network devices managed by [network-operator](https://github.com/ironcore-dev/network-operator). It watches `DeviceMonitor` custom resources and `Device` objects to continuously reconcile the desired monitoring state.

<img :src="architectureImage" alt="Monitoring Operator Architecture" />

## How it works

1. A user creates a **DeviceMonitor** CR specifying which devices to monitor (via label selectors) and which gNMI subscriptions to configure.
2. The **Reconciler** watches both the DeviceMonitor CR and Device resources from network-operator. It reads device endpoint addresses and credentials from the Device specs.
3. The operator creates and manages the following owned resources:
   - **RBAC** (ServiceAccount, Role, RoleBinding) for the gNMIc pods
   - **Secret** containing the generated gNMIc configuration (targets, subscriptions, outputs)
   - **StatefulSet** running gNMIc instances that connect to devices and stream telemetry
   - **Services** for gNMIc clustering and metrics exposure
   - **ServiceMonitor** so Prometheus can discover and scrape the metrics endpoint
4. The gNMIc pods establish **gNMI Subscribe RPCs** to the network devices and expose collected telemetry as Prometheus metrics.
5. **Prometheus Operator** discovers the ServiceMonitor and scrapes `/metrics` from the gNMIc pods, making the data available for dashboards and alerting.
