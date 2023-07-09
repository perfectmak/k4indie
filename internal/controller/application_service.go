package controller

import (
	"context"

	operatorsv1alpha1 "github.com/perfectmak/k4indie/api/v1alpha1"
	"github.com/perfectmak/k4indie/internal/controller/resolvers"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileService attempts to create a service for the application
// if it does not exist. And if it does, it tries to update the service
// schema to match the application spec.
// If the application does not have any endpoints defined, then
// the service will not be created.
func (r *ApplicationReconciler) reconcileService(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
) (*reconcile.Result, error) {
	log := log.FromContext(ctx)

	service := &corev1.Service{}
	err := r.Get(ctx, req.NamespacedName, service)

	noEndpointsDefined := len(appToReconcile.Spec.Endpoints) == 0

	if err != nil && apierrors.IsNotFound(err) {
		if noEndpointsDefined {
			return nil, nil
		}

		service, err = r.createService(ctx, appToReconcile)
		if err != nil {
			log.Error(err, "failed to create service")
			return nil, err
		}

		// Re-enqueue the request to check the status of the service
		return &reconcile.Result{}, nil

	} else if err != nil {
		log.Error(err, "failed to get existing service")
		return nil, err
	}

	if noEndpointsDefined {
		log.Info("deleting service")
		err := r.Delete(ctx, service)
		if err != nil {
			return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
		}

		return nil, nil
	}

	log.Info("updating service")
	err = r.updateServiceSpec(ctx, appToReconcile, service)
	if err != nil {
		return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
	}

	if err := r.Update(ctx, service); err != nil {
		log.Error(err, "failed to update service")

		return r.setApplicationReconcileError(
			ctx, req,
			appToReconcile, log,
			err,
		)
	}

	return nil, nil
}

func (r *ApplicationReconciler) createService(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
) (*corev1.Service, error) {
	log := log.FromContext(ctx)

	service, err := r.buildService(ctx, appToReconcile)
	if err != nil {
		return nil, err
	}

	log.Info(
		"creating service",
		"service.name", service.Name,
		"serfvice.namespace", service.Namespace,
	)

	if err := r.Create(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *ApplicationReconciler) buildService(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
) (*corev1.Service, error) {
	labels := resolvers.MergeDefaultLabels(
		appToReconcile.Labels,
		map[string]string{
			"app.kubernetes.io/instance": appToReconcile.Name,
		})

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appToReconcile.Name,
			Namespace: appToReconcile.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    appToReconcile.Spec.Endpoints.AsServicePorts(),
		},
	}

	if err := ctrl.SetControllerReference(appToReconcile, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *ApplicationReconciler) updateServiceSpec(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
	service *corev1.Service,
) error {
	newService, err := r.buildService(ctx, appToReconcile)
	if err != nil {
		return err
	}

	newService.DeepCopyInto(service)

	return nil
}
