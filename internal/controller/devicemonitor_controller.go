// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"maps"
	"text/template"
	"time"

	networkv1alpha1 "github.com/ironcore-dev/network-operator/api/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/cobaltcore-dev/monitoring-operator/api/v1alpha1"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

// SecretChecksumAnnotation is the annotation key used to store the checksum of the Secret data.
// It is used to trigger a rolling update of the StatefulSet when the Secret data changes.
const SecretChecksumAnnotation = "monitoring.networking.cloud.sap/secret-checksum"

// DeviceMonitorReconciler reconciles a DeviceMonitor object
type DeviceMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Recorder is used to record events for the controller.
	// More info: https://book.kubebuilder.io/reference/raising-events
	Recorder record.EventRecorder

	// RequeueAfter is the duration to wait before requeuing the reconciliation
	// to check for status changes in child resources.
	RequeueAfter time.Duration

	// GNMIcImage is the container image to use for the gNMIc StatefulSet.
	GNMIcImage string
}

// +kubebuilder:rbac:groups=monitoring.networking.cloud.sap,resources=devicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.networking.cloud.sap,resources=devicemonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.networking.cloud.sap,resources=devicemonitors/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// +kubebuilder:rbac:groups=networking.cloud.sap,resources=devices,verbs=get;list;watch

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
//
// For more details about the method shape, read up here:
// - https://ahmet.im/blog/controller-pitfalls/#reconcile-method-shape
func (r *DeviceMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling resource")

	obj := new(v1alpha1.DeviceMonitor)
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("Resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get resource")
		return ctrl.Result{}, err
	}

	if !obj.DeletionTimestamp.IsZero() {
		log.Info("Resource is being deleted, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	if len(obj.Status.Conditions) == 0 {
		log.Info("Initializing status conditions")
		obj.SetReadyCondition(metav1.ConditionUnknown, v1alpha1.ReconcilePendingReason, "Starting reconciliation")
		return ctrl.Result{}, r.Status().Update(ctx, obj)
	}

	// Always attempt to update the status after reconciliation
	orig := obj.DeepCopy()
	defer func() {
		if !equality.Semantic.DeepEqual(orig.Status, obj.Status) {
			if err := r.Status().Patch(ctx, obj, client.MergeFrom(orig)); err != nil {
				log.Error(err, "Failed to update status")
				reterr = kerrors.NewAggregate([]error{reterr, err})
			}
		}
	}()

	if err := r.reconcile(ctx, obj); err != nil {
		log.Error(err, "Failed to reconcile resource")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeviceMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.RequeueAfter == 0 {
		return errors.New("RequeueAfter must be set")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DeviceMonitor{}).
		Named("devicemonitor").
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		// Watches enqueues DeviceMonitors for referenced Device resources.
		Watches(
			&networkv1alpha1.Device{},
			handler.EnqueueRequestsFromMapFunc(r.devicesToDeviceMonitor),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Complete(r)
}

// reconcile handles the main reconciliation logic for the DeviceMonitor resource.
//
// It performs the following steps:
// 1. Create a ServiceAccount for the gnmic StatefulSet.
// 2. Create a Role with the necessary permissions for the gnmic StatefulSet.
// 3. Create a RoleBinding to bind the Role to the ServiceAccount.
// 4. Create a Secret to store the gnmic configuration.
// 5. Create a StatefulSet running gnmic to collect metrics from the devices.
// 6. Create a Service to expose the gnmic-api (for clustering) of the gnmic StatefulSet.
// 7. Create a Service to expose the prometheus metrics collected by the gnmic StatefulSet.
// 8. Create a ServiceMonitor to scrape metrics from the Service.
func (r *DeviceMonitorReconciler) reconcile(ctx context.Context, m *v1alpha1.DeviceMonitor) (reterr error) {
	log := ctrl.LoggerFrom(ctx)

	labels := map[string]string{
		"app.kubernetes.io/name":       m.Name,
		"app.kubernetes.io/instance":   m.Name + "-" + string(m.UID),
		"app.kubernetes.io/managed-by": "monitoring-operator",
		v1alpha1.DeviceMonitorLabel:    m.Name,
	}

	selector, err := metav1.LabelSelectorAsSelector(&m.Spec.Selector)
	if err != nil {
		log.Error(err, "Failed to parse device selector")
		return err
	}

	var devices networkv1alpha1.DeviceList
	if err := r.List(ctx, &devices, &client.ListOptions{LabelSelector: client.MatchingLabelsSelector{Selector: selector}}); err != nil {
		log.Error(err, "Failed to list Devices")
		return err
	}

	if len(devices.Items) == 0 {
		log.Info(fmt.Sprintf("No devices found matching selector %v", m.Spec.Selector))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "NoDevices", "No devices found matching selector %v", m.Spec.Selector)

		// Cleanup any existing resources since there are no devices to monitor
		obj := []client.Object{
			&corev1.ServiceAccount{},
			&rbacv1.Role{},
			&rbacv1.RoleBinding{},
			&corev1.Secret{},
			&appsv1.StatefulSet{},
			&corev1.Service{},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: m.ServiceName()}},
			&monitoringv1.ServiceMonitor{},
		}
		var errs []error
		for _, o := range obj {
			if o.GetName() == "" {
				o.SetName(m.Name)
			}
			o.SetNamespace(m.Namespace)

			if err := r.Delete(ctx, o); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Error(err, "Failed to delete resource", "Kind", o.GetObjectKind().GroupVersionKind().Kind)
					errs = append(errs, err)
				}
				continue
			}

			r.Recorder.Eventf(m, corev1.EventTypeNormal, "Deleted", "Deleted %s %s/%s", o.GetObjectKind().GroupVersionKind().Kind, o.GetNamespace(), o.GetName())
		}
		if len(errs) > 0 {
			return kerrors.NewAggregate(errs)
		}

		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("No devices found matching selector %v", m.Spec.Selector))
	}

	targets := make([]Target, 0, len(devices.Items))
	for _, d := range devices.Items {
		user, pass, err := r.BasicAuth(ctx, d.Namespace, d.Spec.Endpoint.SecretRef)
		if err != nil {
			log.Error(err, "Failed to get basic auth credentials for device", "Device", fmt.Sprintf("%s/%s", d.Namespace, d.Name))
			return err
		}

		targets = append(targets, Target{
			Name:     d.Name,
			Address:  d.Spec.Endpoint.Address,
			Username: string(user),
			Password: string(pass),
		})
	}

	sa := &corev1.ServiceAccount{}
	sa.Name = m.Name
	sa.Namespace = m.Namespace
	res, err := controllerutil.CreateOrPatch(ctx, r.Client, sa, func() error {
		ensureLabels(sa, labels)
		return controllerutil.SetControllerReference(m, sa, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update ServiceAccount", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update ServiceAccount: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "ServiceAccount %s/%s: %v", sa.Namespace, sa.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "ServiceAccount %s/%s %s", sa.Namespace, sa.Name, res)
	}

	ro := &rbacv1.Role{}
	ro.Name = m.Name
	ro.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, ro, func() error {
		ensureLabels(ro, labels)
		ro.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		}
		return controllerutil.SetControllerReference(m, ro, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update Role", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update Role: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "Role %s/%s: %v", ro.Namespace, ro.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "Role %s/%s %s", ro.Namespace, ro.Name, res)
	}

	rb := &rbacv1.RoleBinding{}
	rb.Name = m.Name
	rb.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, rb, func() error {
		ensureLabels(rb, labels)
		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     ro.Name,
		}
		rb.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		}
		return controllerutil.SetControllerReference(m, rb, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update RoleBinding", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update RoleBinding: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "RoleBinding %s/%s: %v", rb.Namespace, rb.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "RoleBinding %s/%s %s", rb.Namespace, rb.Name, res)
	}

	// checksum of the gnmic config.yaml data stored in the Secret.
	// If the config changes, the checksum will change and trigger a rolling update of the StatefulSet.
	var checksum string

	s := &corev1.Secret{}
	s.Name = m.Name
	s.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, s, func() error {
		ensureLabels(s, labels)
		s.Type = corev1.SecretTypeOpaque
		if s.Data == nil {
			s.Data = map[string][]byte{}
		}
		var buf bytes.Buffer
		err := templates.ExecuteTemplate(&buf, "config.tmpl", map[string]any{
			"Name":          m.Name,
			"Namespace":     m.Namespace,
			"Subscriptions": m.Spec.Subscriptions,
			"Targets":       targets,
		})
		if err != nil {
			return err
		}
		data := buf.Bytes()
		s.Data["config.yaml"] = data
		checksum = fmt.Sprintf("%x", sha256.Sum256(data))
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}
		s.Annotations[SecretChecksumAnnotation] = checksum
		return controllerutil.SetControllerReference(m, s, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update Secret", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update Secret: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "Secret %s/%s: %v", s.Namespace, s.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "Secret %s/%s %s", s.Namespace, s.Name, res)
	}

	sts := &appsv1.StatefulSet{}
	sts.Name = m.Name
	sts.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, sts, func() error {
		ensureLabels(sts, labels)
		sts.Spec.Replicas = &m.Spec.Replicas
		sts.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		sts.Spec.ServiceName = m.Name
		m.Spec.Template.ObjectMeta.DeepCopyInto(&sts.Spec.Template.ObjectMeta)
		// Add the secret checksum annotation to the pod template to trigger a rolling update when the secret changes.
		if sts.Spec.Template.Annotations == nil {
			sts.Spec.Template.Annotations = map[string]string{}
		}
		sts.Spec.Template.Annotations[SecretChecksumAnnotation] = checksum
		ensureLabels(&sts.Spec.Template, labels)
		if sts.Spec.Template.Spec.NodeSelector == nil {
			sts.Spec.Template.Spec.NodeSelector = m.Spec.Template.Spec.NodeSelector
		}
		if m.Spec.Template.Spec.Affinity != nil {
			sts.Spec.Template.Spec.Affinity = m.Spec.Template.Spec.Affinity
		}
		if len(sts.Spec.Template.Spec.Containers) != 1 {
			sts.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
		}
		sts.Spec.Template.Spec.Containers[0].Name = "gnmic"
		sts.Spec.Template.Spec.Containers[0].Image = r.GNMIcImage
		sts.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
		sts.Spec.Template.Spec.Containers[0].Args = []string{"subscribe", "--config", "/etc/gnmic/config.yaml"}
		sts.Spec.Template.Spec.ServiceAccountName = sa.Name
		sts.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			AllowPrivilegeEscalation: ptr.To(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			ReadOnlyRootFilesystem: ptr.To(true),
			RunAsNonRoot:           ptr.To(true),
			RunAsUser:              ptr.To(int64(1000)),
		}
		sts.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "prometheus",
				ContainerPort: 9804,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "gnmic-api",
				ContainerPort: 7890,
				Protocol:      corev1.ProtocolTCP,
			},
		}
		sts.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("150m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		}
		sts.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
			{
				Name:  "GNMIC_API",
				Value: ":7890",
			},
			{
				Name: "GNMIC_CLUSTERING_INSTANCE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			{
				Name: "GNMIC_CLUSTERING_NAMESPACE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},
			{
				Name:  "GNMIC_CLUSTERING_SERVICE_NAME",
				Value: m.Name,
			},
			{
				Name:  "GNMIC_CLUSTERING_SERVICE_ADDRESS",
				Value: "$(GNMIC_CLUSTERING_INSTANCE_NAME).$(GNMIC_CLUSTERING_SERVICE_NAME).$(GNMIC_CLUSTERING_NAMESPACE_NAME).svc.cluster.local",
			},
			{
				Name:  "GNMIC_OUTPUTS_PROM_LISTEN",
				Value: "$(GNMIC_CLUSTERING_INSTANCE_NAME).$(GNMIC_CLUSTERING_SERVICE_NAME).$(GNMIC_CLUSTERING_NAMESPACE_NAME).svc.cluster.local:9804",
			},
		}
		sts.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/etc/gnmic/config.yaml",
				SubPath:   "config.yaml",
				ReadOnly:  true,
			},
		}
		sts.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  s.Name,
						DefaultMode: ptr.To[int32](0o440),
					},
				},
			},
		}
		return controllerutil.SetControllerReference(m, sts, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update StatefulSet", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update StatefulSet: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "StatefulSet %s/%s: %v", sts.Namespace, sts.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "StatefulSet %s/%s %s", sts.Namespace, sts.Name, res)
	}

	// GNMI expects a dedicated service that only has a single port 7890/TCP for the clustering API.
	// See: https://github.com/openconfig/gnmic/blob/0aa04b5727cd894ef7ef9e8f787dc2f9c513e5b0/pkg/lockers/k8s_locker/k8s_registration.go#L118-L119
	svc := &corev1.Service{}
	svc.Name = m.ServiceName()
	svc.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, svc, func() error {
		ensureLabels(svc, labels)
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       7890,
				TargetPort: intstr.FromInt(7890),
				Protocol:   corev1.ProtocolTCP,
			},
		}
		svc.Spec.ClusterIP = corev1.ClusterIPNone // Headless Service for StatefulSet
		return controllerutil.SetControllerReference(m, svc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update Service", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update Service: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "Service %s/%s: %v", svc.Namespace, svc.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "Service %s/%s %s", svc.Namespace, svc.Name, res)
	}

	svc = &corev1.Service{}
	svc.Name = m.Name
	svc.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, svc, func() error {
		ensureLabels(svc, labels)
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       9804,
				TargetPort: intstr.FromInt(9804),
				Protocol:   corev1.ProtocolTCP,
			},
		}
		svc.Spec.ClusterIP = corev1.ClusterIPNone // Headless Service for StatefulSet
		return controllerutil.SetControllerReference(m, svc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update Service", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update Service: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "Service %s/%s: %v", svc.Namespace, svc.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "Service %s/%s %s", svc.Namespace, svc.Name, res)
	}

	sm := &monitoringv1.ServiceMonitor{}
	sm.Name = m.Name
	sm.Namespace = m.Namespace
	res, err = controllerutil.CreateOrPatch(ctx, r.Client, sm, func() error {
		ensureLabels(sm, labels)
		sm.Spec.Selector = metav1.LabelSelector{MatchLabels: labels}
		sm.Spec.Endpoints = []monitoringv1.Endpoint{
			{
				Port: "http",
				Path: "/metrics",
			},
		}
		sm.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
			MatchNames: []string{m.Namespace},
		}
		return controllerutil.SetControllerReference(m, sm, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create or update ServiceMonitor", "result", res)
		m.SetReadyCondition(metav1.ConditionFalse, v1alpha1.NotReadyReason, fmt.Sprintf("Failed to create or update ServiceMonitor: %v", err))
		r.Recorder.Eventf(m, corev1.EventTypeWarning, "ReconcileError", "ServiceMonitor %s/%s: %v", sm.Namespace, sm.Name, err)
		return err
	}
	if res != controllerutil.OperationResultNone {
		r.Recorder.Eventf(m, corev1.EventTypeNormal, "Reconciled", "ServiceMonitor %s/%s %s", sm.Namespace, sm.Name, res)
	}

	m.SetReadyCondition(metav1.ConditionTrue, v1alpha1.ReadyCondition, "All owned resources are ready")
	return nil
}

// BasicAuth loads the username and password from the referenced secret resource.
// The secret must by of type 'kubernetes.io/basic-auth' and contain the fields 'username' and 'password'.
func (r *DeviceMonitorReconciler) BasicAuth(ctx context.Context, defaultNamespace string, ref *networkv1alpha1.SecretReference) (user, pass []byte, err error) {
	key := client.ObjectKey{Namespace: ref.Namespace, Name: ref.Name}
	if key.Namespace == "" {
		key.Namespace = defaultNamespace
	}

	var secret corev1.Secret
	if err := r.Get(ctx, key, &secret); err != nil {
		return nil, nil, fmt.Errorf("failed to get secret %q: %w", key.String(), err)
	}

	if secret.Type != corev1.SecretTypeBasicAuth {
		return nil, nil, fmt.Errorf("unsupported secret type: want %q, got %q", corev1.SecretTypeBasicAuth, secret.Type)
	}

	user, ok := secret.Data[corev1.BasicAuthUsernameKey]
	if !ok || len(user) == 0 {
		return nil, nil, fmt.Errorf("missing field 'username' in secret %q", key.String())
	}

	pass, ok = secret.Data[corev1.BasicAuthPasswordKey]
	if !ok || len(pass) == 0 {
		return nil, nil, fmt.Errorf("missing field 'password' in secret %q", key.String())
	}

	return user, pass, nil
}

// devicesToDeviceMonitor is a [handler.MapFunc] to be used to enqueue requests for reconciliation
// for a DeviceMonitor to update when one of its referenced Devices gets updated.
func (r *DeviceMonitorReconciler) devicesToDeviceMonitor(ctx context.Context, obj client.Object) []ctrl.Request {
	dev, ok := obj.(*networkv1alpha1.Device)
	if !ok {
		panic(fmt.Sprintf("Expected a Device but got a %T", obj))
	}

	log := ctrl.LoggerFrom(ctx, "Device", klog.KObj(dev))

	monitors := new(v1alpha1.DeviceMonitorList)
	if err := r.List(ctx, monitors); err != nil {
		log.Error(err, "Failed to list DeviceMonitors")
		return nil
	}

	requests := []ctrl.Request{}
	for _, m := range monitors.Items {
		selector, err := metav1.LabelSelectorAsSelector(&m.Spec.Selector)
		if err != nil {
			log.Error(err, "Failed to parse device selector", "DeviceMonitor", klog.KObj(&m))
			continue
		}

		if selector.Matches(klabels.Set(dev.Labels)) {
			log.Info("Enqueuing DeviceMonitor for reconciliation", "DeviceMonitor", klog.KObj(&m))
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
			})
		}
	}

	return requests
}

// ensureLabels ensures that the given labels are present on the object.
// It adds the labels if they are not already present while preserving existing labels.
func ensureLabels(obj metav1.Object, labels map[string]string) {
	if obj.GetLabels() == nil {
		obj.SetLabels(make(map[string]string))
	}
	maps.Copy(obj.GetLabels(), labels)
}

// Target represents a network device to be monitored. It includes the necessary gnmic target configuration.
// See: https://gnmic.openconfig.net/user_guide/targets/targets
type Target struct {
	Name     string
	Address  string
	Username string
	Password string
}
