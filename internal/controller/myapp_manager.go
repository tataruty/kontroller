package controller

import (
	"context"
	"fmt"
	"time"

	"honnef.co/go/tools/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctlr "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// clusterTimeout is a timeout for connections to the Kubernetes API.
	clusterTimeout = 10 * time.Second
)

func createManager(cfg config.Config, options manager.Options) (manager.Manager, error) {
	if cfg.HealthConfig.Enabled {
		options.HealthProbeBindAddress = fmt.Sprintf(":%d", cfg.HealthConfig.Port)
	}

	clusterCfg := ctlr.GetConfigOrDie()
	clusterCfg.Timeout = clusterTimeout

	mgr, err := manager.New(clusterCfg, options)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func registerControllers(
	ctx context.Context,
	cfg config.Config,
	mgr manager.Manager,
	recorder record.EventRecorder,
	logLevelSetter logLevelSetter,
	eventCh chan interface{},
	controlConfigNSName types.NamespacedName,
) error {
	type ctlrCfg struct {
		name       string
		objectType ngftypes.ObjectType
		options    []controller.Option
	}

	crdWithGVK := apiext.CustomResourceDefinition{}
	crdWithGVK.SetGroupVersionKind(
		schema.GroupVersionKind{Group: apiext.GroupName, Version: "v1", Kind: "CustomResourceDefinition"},
	)

	// Note: for any new object type or a change to the existing one,
	// make sure to also update prepareFirstEventBatchPreparerArgs()
	controllerRegCfgs := []ctlrCfg{
		{
			objectType: &gatewayv1.GatewayClass{},
			options: []controller.Option{
				controller.WithK8sPredicate(
					k8spredicate.And(
						k8spredicate.GenerationChangedPredicate{},
						predicate.GatewayClassPredicate{ControllerName: cfg.GatewayCtlrName},
					),
				),
			},
		},
		{
			objectType: &gatewayv1.Gateway{},
			options: func() []controller.Option {
				options := []controller.Option{
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				}
				return options
			}(),
		},
		{
			objectType: &gatewayv1.HTTPRoute{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &apiv1.Service{},
			name:       "user-service", // unique controller names are needed and we have multiple Service ctlrs
			options: []controller.Option{
				controller.WithK8sPredicate(predicate.ServiceChangedPredicate{}),
			},
		},
		{
			objectType: &apiv1.Secret{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.ResourceVersionChangedPredicate{}),
			},
		},
		{
			objectType: &discoveryV1.EndpointSlice{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				controller.WithFieldIndices(index.CreateEndpointSliceFieldIndices()),
			},
		},
		{
			objectType: &apiv1.Namespace{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.LabelChangedPredicate{}),
			},
		},
		{
			objectType: &gatewayv1beta1.ReferenceGrant{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &crdWithGVK,
			options: []controller.Option{
				controller.WithOnlyMetadata(),
				controller.WithK8sPredicate(
					predicate.AnnotationPredicate{Annotation: graph.BundleVersionAnnotation},
				),
			},
		},
		{
			objectType: &ngfAPIv1alpha2.NginxProxy{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &gatewayv1.GRPCRoute{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &ngfAPIv1alpha1.ClientSettingsPolicy{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &ngfAPIv1alpha2.ObservabilityPolicy{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
		{
			objectType: &ngfAPIv1alpha1.UpstreamSettingsPolicy{},
			options: []controller.Option{
				controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
			},
		},
	}

	if cfg.ExperimentalFeatures {
		gwExpFeatures := []ctlrCfg{
			{
				objectType: &gatewayv1alpha3.BackendTLSPolicy{},
				options: []controller.Option{
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				},
			},
			{
				// FIXME(ciarams87): If possible, use only metadata predicate
				// https://github.com/nginx/nginx-gateway-fabric/issues/1545
				objectType: &apiv1.ConfigMap{},
			},
			{
				objectType: &gatewayv1alpha2.TLSRoute{},
				options: []controller.Option{
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				},
			},
		}
		controllerRegCfgs = append(controllerRegCfgs, gwExpFeatures...)
	}

	if cfg.ConfigName != "" {
		controllerRegCfgs = append(controllerRegCfgs,
			ctlrCfg{
				objectType: &ngfAPIv1alpha1.NginxGateway{},
				options: []controller.Option{
					controller.WithNamespacedNameFilter(filter.CreateSingleResourceFilter(controlConfigNSName)),
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				},
			})
		if err := setInitialConfig(
			mgr.GetAPIReader(),
			cfg.Logger,
			recorder,
			logLevelSetter,
			controlConfigNSName,
		); err != nil {
			return fmt.Errorf("error setting initial control plane configuration: %w", err)
		}
	}

	if cfg.SnippetsFilters {
		controllerRegCfgs = append(controllerRegCfgs,
			ctlrCfg{
				objectType: &ngfAPIv1alpha1.SnippetsFilter{},
				options: []controller.Option{
					controller.WithK8sPredicate(k8spredicate.GenerationChangedPredicate{}),
				},
			},
		)
	}

	for _, regCfg := range controllerRegCfgs {
		name := regCfg.objectType.GetObjectKind().GroupVersionKind().Kind
		if regCfg.name != "" {
			name = regCfg.name
		}

		if err := controller.Register(
			ctx,
			regCfg.objectType,
			name,
			mgr,
			eventCh,
			regCfg.options...,
		); err != nil {
			return fmt.Errorf("cannot register controller for %T: %w", regCfg.objectType, err)
		}
	}
	return nil
}
