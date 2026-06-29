// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cobaltcore-dev/gnmi-exporter/api/v1alpha1"
)

var _ = Describe("DeviceMonitor Controller", func() {
	Context("When reconciling a resource", func() {
		const name = "test-monitor"
		key := client.ObjectKey{Name: name, Namespace: metav1.NamespaceDefault}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind DeviceMonitor")
			m := &v1alpha1.DeviceMonitor{}
			if err := k8sClient.Get(ctx, key, m); apierrors.IsNotFound(err) {
				m = &v1alpha1.DeviceMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: metav1.NamespaceDefault,
					},
					Spec: v1alpha1.DeviceMonitorSpec{
						Replicas: 3,
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{"topology.kubernetes.io/zone": "eu-de-1a"},
						},
						Template: v1alpha1.PodTemplateSpec{
							Spec: v1alpha1.PodSpec{
								NodeSelector: map[string]string{"topology.kubernetes.io/zone": "eu-de-1a"},
							},
						},
						Subscriptions: []v1alpha1.Subscription{
							{
								Name:  "interfaces",
								Paths: []string{"/interfaces/interface/state/counters"},
							},
						},
						Encoding: v1alpha1.EncodingJSONIETF,
						Metrics: &v1alpha1.MetricsSpec{
							ServiceMonitor: &v1alpha1.ServiceMonitorSpec{
								AdditionalLabels: map[string]string{"prometheus": "infra"},
								Interval:         "30s",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, m)).To(Succeed())
			}
		})

		AfterEach(func() {
			m := &v1alpha1.DeviceMonitor{}
			err := k8sClient.Get(ctx, key, m)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DeviceMonitor")
			Expect(k8sClient.Delete(ctx, m)).To(Succeed())

			// TODO: Add specific tests to verify the deletion, e.g. resource cleanup
		})

		It("Should successfully reconcile the resource", func() {
			// TODO: Add specific tests to verify the reconciliation, e.g. resource creation
		})
	})
})
