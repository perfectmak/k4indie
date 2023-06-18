package controller

import (
	"context"

	operatorsv1alpha1 "github.com/perfectmak/k4indie/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileService(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
) (*reconcile.Result, error) {
	// log := log.FromContext(ctx)

	service := &corev1.Service{}
	err := r.Get(ctx, req.NamespacedName, service)

	return nil, err
}
