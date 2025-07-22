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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	webappv1 "my-apps.com/myapp/api/v1"
)

// MyappReconciler reconciles a Myapp object
type MyappReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.my-apps.com,resources=myapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.my-apps.com,resources=myapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.my-apps.com,resources=myapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Myapp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *MyappReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	logger.Info("------------------------- Reconciling the resource -------------------------\n")

	// Check if this is an HTTPRoute event
	var httpRoute gatewayv1.HTTPRoute
	if err := r.Get(ctx, req.NamespacedName, &httpRoute); err == nil {
		logger.Info("HTTPRoute event detected",
			"name", httpRoute.Name,
			"namespace", httpRoute.Namespace,
			"generation", httpRoute.Generation,
			"resourceVersion", httpRoute.ResourceVersion)

		// List all HTTPRoutes in the same namespace
		var httpRoutes gatewayv1.HTTPRouteList
		if err := r.List(ctx, &httpRoutes, client.InNamespace(req.Namespace)); err != nil {
			logger.Info("Failed to list HTTPRoutes (Gateway API may not be available)", "error", err)
		} else {
			logger.Info("<======== All HTTPRoutes in namespace =======>\n", "count", len(httpRoutes.Items), "namespace", req.Namespace)
		}
		return ctrl.Result{}, nil
	}

	// List all HTTPRoutes in the same namespace to detect changes
	var httpRoutes gatewayv1.HTTPRouteList
	if err := r.List(ctx, &httpRoutes, client.InNamespace(req.Namespace)); err != nil {
		logger.Info("Failed to list HTTPRoutes (Gateway API may not be available)", "error", err)
		// Continue with Myapp processing even if HTTPRoute listing fails
	} else {
		logger.Info("Found HTTPRoutes in namespace", "count", len(httpRoutes.Items), "namespace", req.Namespace)

		// Log each HTTPRoute for debugging
		for _, route := range httpRoutes.Items {
			logger.Info("HTTPRoute in namespace",
				"name", route.Name,
				"namespace", route.Namespace,
				"generation", route.Generation,
				"resourceVersion", route.ResourceVersion)
		}
	}

	// Check if this is an Service event
	var service corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &service); err == nil {
		logger.Info("Service event detected",
			"name", service.Name,
			"namespace", service.Namespace,
			"generation", service.Generation,
			"resourceVersion", service.ResourceVersion)

		// List all Services in the same namespace
		var services corev1.ServiceList
		if err := r.List(ctx, &services, client.InNamespace(req.Namespace)); err != nil {
			logger.Info("Failed to list Services", "error", err)
		} else {
			logger.Info("<======== All Services in namespace =======>\n", "count", len(services.Items), "namespace", req.Namespace)
		}
		return ctrl.Result{}, nil
	}

	// List all Services in the same namespace to detect changes
	var services corev1.ServiceList
	if err := r.List(ctx, &services, client.InNamespace(req.Namespace)); err != nil {
		logger.Info("Failed to list Services", "error", err)
		// Continue with Myapp processing even if Services listing fails
	} else {
		logger.Info("Found Services in namespace", "count", len(services.Items), "namespace", req.Namespace)

		// Log each Service for debugging
		for _, service := range services.Items {
			logger.Info("Service in namespace",
				"name", service.Name,
				"namespace", service.Namespace,
				"generation", service.Generation,
				"resourceVersion", service.ResourceVersion)
		}
	}

	// Try to fetch as Myapp instance
	var myapp webappv1.Myapp
	if err := r.Get(ctx, req.NamespacedName, &myapp); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Resource not found. Ignoring since object must be deleted", "namespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get resource")
		return ctrl.Result{}, err
	}

	logger.Info("Myapp event detected", "name", myapp.Name, "namespace", myapp.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyappReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrl.Log.Info("Setting up controller with the manager")
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Myapp{}).
		Watches(&gatewayv1.HTTPRoute{}, &handler.EnqueueRequestForObject{}).
		Watches(&corev1.Service{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
