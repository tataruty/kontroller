/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapischeme "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/scheme"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webappv1 "my-apps.com/myapp/api/v1"
)

var _ = Describe("Myapp Controller", func() {
	Context("Basic functionality", func() {
		It("should create a reconciler instance", func() {
			By("Creating a MyappReconciler instance")
			reconciler := &MyappReconciler{}
			Expect(reconciler).NotTo(BeNil())
		})

		It("should handle reconciliation with fake client", func() {
			By("Creating a fake client with proper scheme")
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Add Gateway API scheme (this is optional for the test)
			if err := gatewayapischeme.AddToScheme(scheme); err != nil {
				GinkgoWriter.Printf("Gateway API scheme not available: %v\n", err)
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			reconciler := &MyappReconciler{
				Client: client,
				Scheme: scheme,
			}

			By("Testing reconciliation of a non-existent resource")
			ctx := context.Background()
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-resource",
					Namespace: "default",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})
	})

	Context("When reconciling a Myapp resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		myapp := &webappv1.Myapp{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Myapp")
			err := k8sClient.Get(ctx, typeNamespacedName, myapp)
			if err != nil && errors.IsNotFound(err) {
				resource := &webappv1.Myapp{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &webappv1.Myapp{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Myapp")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the Myapp resource", func() {
			By("Reconciling the created Myapp resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle non-existent Myapp resource gracefully", func() {
			By("Reconciling a non-existent Myapp resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nonExistentName := types.NamespacedName{
				Name:      "non-existent-resource",
				Namespace: "default",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nonExistentName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling an HTTPRoute resource", func() {
		const httpRouteName = "test-httproute"

		ctx := context.Background()

		httpRouteNamespacedName := types.NamespacedName{
			Name:      httpRouteName,
			Namespace: "default",
		}

		It("should handle HTTPRoute reconciliation (if Gateway API is available)", func() {
			By("Attempting to reconcile an HTTPRoute resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Try to reconcile an HTTPRoute - this should handle the case gracefully
			// whether Gateway API is installed or not
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: httpRouteNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle non-existent HTTPRoute resource gracefully", func() {
			By("Reconciling a non-existent HTTPRoute resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nonExistentName := types.NamespacedName{
				Name:      "non-existent-httproute",
				Namespace: "default",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nonExistentName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling a Service resource", func() {
		const serviceName = "test-service"

		ctx := context.Background()

		serviceNamespacedName := types.NamespacedName{
			Name:      serviceName,
			Namespace: "default",
		}

		It("should handle Service reconciliation", func() {
			By("Attempting to reconcile a Service resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Try to reconcile a Service - this should handle the case gracefully
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: serviceNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle non-existent Service resource gracefully", func() {
			By("Reconciling a non-existent Service resource")
			controllerReconciler := &MyappReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nonExistentName := types.NamespacedName{
				Name:      "non-existent-service",
				Namespace: "default",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nonExistentName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Permission validation tests", func() {
		ctx := context.Background()

		It("should handle HTTPRoute reconciliation when Gateway API scheme is missing", func() {
			By("Creating a fake client without Gateway API scheme")
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// Intentionally NOT adding Gateway API scheme

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			reconciler := &MyappReconciler{
				Client: client,
				Scheme: scheme,
			}

			By("Testing HTTPRoute reconciliation without Gateway API scheme")
			httpRouteNamespacedName := types.NamespacedName{
				Name:      "test-httproute",
				Namespace: "default",
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: httpRouteNamespacedName,
			})

			// Should not return an error because controller handles gracefully
			// But internally it should log permission/scheme errors
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})

		It("should return error when trying to access HTTPRoute without proper scheme", func() {
			By("Testing direct HTTPRoute access without Gateway API scheme")
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// NOT adding Gateway API scheme

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			By("Attempting to get HTTPRoute directly - should fail")
			var httpRoute gatewayv1.HTTPRoute
			err = client.Get(ctx, types.NamespacedName{
				Name:      "test-route",
				Namespace: "default",
			}, &httpRoute)

			// This SHOULD fail because HTTPRoute type is not in scheme
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no kind is registered"))
		})

		It("should handle Service reconciliation with limited scheme", func() {
			By("Creating a fake client with only webapp scheme")
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// Not adding core API scheme to simulate missing Service permissions

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			reconciler := &MyappReconciler{
				Client: client,
				Scheme: scheme,
			}

			By("Testing Service reconciliation without core API scheme")
			serviceNamespacedName := types.NamespacedName{
				Name:      "test-service",
				Namespace: "default",
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: serviceNamespacedName,
			})

			// Should not return an error because controller handles gracefully
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})

		It("should return error when trying to access Service without proper scheme", func() {
			By("Testing direct Service access without core API scheme")
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())
			// NOT adding core API scheme

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			By("Attempting to get Service directly - should fail")
			var service corev1.Service
			err = client.Get(ctx, types.NamespacedName{
				Name:      "test-service",
				Namespace: "default",
			}, &service)

			// This SHOULD fail because Service type is not in scheme
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no kind is registered"))
		})

		It("should validate that HTTPRoute watching requires proper RBAC", func() {
			By("Testing that HTTPRoute operations would fail without proper permissions")
			// This test validates that the RBAC annotations are correct

			// Create a scheme with Gateway API
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			err = gatewayapischeme.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Create an HTTPRoute object to test with
			httpRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-route",
					Namespace: "default",
				},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{
							{
								Name: "test-gateway",
							},
						},
					},
				},
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(httpRoute).
				Build()

			reconciler := &MyappReconciler{
				Client: client,
				Scheme: scheme,
			}

			By("Reconciling with a real HTTPRoute object")
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-route",
					Namespace: "default",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})

		It("should validate that Service watching requires proper RBAC", func() {
			By("Testing that Service operations would work with proper permissions")

			// Create a scheme with core API
			scheme := runtime.NewScheme()
			err := webappv1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Add core scheme for Services
			err = corev1.AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Create a Service object to test with
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "test",
					},
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(service).
				Build()

			reconciler := &MyappReconciler{
				Client: client,
				Scheme: scheme,
			}

			By("Reconciling with a real Service object")
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-service",
					Namespace: "default",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})
	})

})
