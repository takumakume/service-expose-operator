package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	serviceexposev1alpha1 "github.com/takumakume/service-expose-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ServiceExpose controller", func() {
	const (
		serviceExposeName = "test-serviceexpose"

		serviceName       = "test-service"
		servicePort int32 = 80

		timeout  = time.Second * 60
		duration = time.Second * 30
		interval = time.Second * 1
	)

	Context("When updating ServiceExpose Status", func() {
		It("Should generate Ingress", func() {
			ctx := context.Background()

			By("By creating Service")
			svc := &v1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{Port: servicePort},
					},
				},
			}
			Expect(k8sClient.Create(ctx, svc)).Should(Succeed())

			By("By creating ServiceExpose")
			seviceexpose := &serviceexposev1alpha1.ServiceExpose{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "service-expose.takumakume.github.io/v1alpha1",
					Kind:       "ServiceExpose",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceExposeName,
					Namespace: "default",
				},
				Spec: serviceexposev1alpha1.ServiceExposeSpec{
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: serviceName,
							Port: networkingv1.ServiceBackendPort{
								Number: 80,
							},
						},
					},
					Path:     "/",
					PathType: networkingv1.PathTypePrefix,
					Domain:   "example.com",
				},
			}
			Expect(k8sClient.Create(ctx, seviceexpose)).Should(Succeed())

			By("By checking exists ServiceExpose")
			serviceExposeLookupKey := types.NamespacedName{Name: serviceExposeName, Namespace: "default"}
			createdServiceExpose := &serviceexposev1alpha1.ServiceExpose{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceExposeLookupKey, createdServiceExpose)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdServiceExpose.Spec.Path).Should(Equal("/"))

			By("By checking ServiceExpose is ready")
			Eventually(func() (v1.ConditionStatus, error) {
				err := k8sClient.Get(ctx, serviceExposeLookupKey, createdServiceExpose)
				if err != nil {
					return "", err
				}
				return createdServiceExpose.Status.Ready, nil
			}, duration, interval).Should(Equal(v1.ConditionTrue))
			Expect(createdServiceExpose.Status.IngressName).Should(Equal("test-serviceexpose-ingress"))
			Expect(createdServiceExpose.Status.IngressHost).Should(Equal("test-service.default.example.com"))

			By("By checking generated Ingress")
			ingressLookupKey := types.NamespacedName{Name: createdServiceExpose.Status.IngressName, Namespace: "default"}
			createdIngress := &networkingv1.Ingress{}
			Expect(k8sClient.Get(ctx, ingressLookupKey, createdIngress)).Should(Succeed())
			Expect(len(createdIngress.Spec.Rules)).Should(Equal(1))
			Expect(len(createdIngress.Spec.TLS)).Should(Equal(0))
			Expect(createdIngress.Spec.Rules[0].Host).Should(Equal("test-service.default.example.com"))
			Expect(len(createdIngress.Spec.Rules[0].HTTP.Paths)).Should(Equal(1))
			Expect(createdIngress.Spec.Rules[0].HTTP.Paths[0].Path).Should(Equal("/"))
			Expect(*createdIngress.Spec.Rules[0].HTTP.Paths[0].PathType).Should(Equal(networkingv1.PathTypePrefix))
			Expect(createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).Should(Equal(serviceName))
			Expect(createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).Should(Equal(servicePort))
		})
	})

})
